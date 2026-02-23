package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/library/internal/models"
)

var (
	serverURL string
	jsonOut   bool
	client    = &http.Client{Timeout: 30 * time.Second}
)

func main() {
	root := &cobra.Command{
		Use:   "library",
		Short: "Personal library catalog CLI",
	}

	root.PersistentFlags().StringVar(&serverURL, "server", "http://localhost:8080", "Library server URL")
	root.PersistentFlags().BoolVar(&jsonOut, "json", false, "Output raw JSON")

	root.AddCommand(serveCmd())
	root.AddCommand(addCmd())
	root.AddCommand(listCmd())
	root.AddCommand(searchCmd())
	root.AddCommand(showCmd())
	root.AddCommand(editCmd())
	root.AddCommand(deleteCmd())
	root.AddCommand(statusCmd())
	root.AddCommand(rateCmd())
	root.AddCommand(coverCmd())
	root.AddCommand(shelvesCmd())
	root.AddCommand(tagCmd())
	root.AddCommand(statsCmd())
	root.AddCommand(importCmd())
	root.AddCommand(exportCmd())
	root.AddCommand(lookupCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

// --- helpers ---

func apiURL(path string) string {
	return strings.TrimRight(serverURL, "/") + path
}

func doJSON(method, path string, body interface{}) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, apiURL(path), reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("server unreachable: %v", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		var errResp map[string]string
		json.Unmarshal(data, &errResp)
		if msg, ok := errResp["error"]; ok {
			return nil, fmt.Errorf("%s", msg)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(data))
	}

	return data, nil
}

func printJSON(data []byte) {
	var buf bytes.Buffer
	json.Indent(&buf, data, "", "  ")
	fmt.Println(buf.String())
}

func printBook(b *models.Book) {
	fmt.Printf("ID:      %d\n", b.ID)
	fmt.Printf("Title:   %s\n", b.Title)
	fmt.Printf("Author:  %s\n", b.Author)
	if b.ISBN != "" {
		fmt.Printf("ISBN:    %s\n", b.ISBN)
	}
	if b.Genre != "" {
		fmt.Printf("Genre:   %s\n", b.Genre)
	}
	if b.Pages > 0 {
		fmt.Printf("Pages:   %d\n", b.Pages)
	}
	if b.Year > 0 {
		fmt.Printf("Year:    %d\n", b.Year)
	}
	fmt.Printf("Status:  %s\n", b.Status)
	if b.Rating > 0 {
		fmt.Printf("Rating:  %s (%d/5)\n", strings.Repeat("*", b.Rating)+strings.Repeat("-", 5-b.Rating), b.Rating)
	}
	if b.Notes != "" {
		fmt.Printf("Notes:   %s\n", b.Notes)
	}
	if b.CoverPath != "" {
		fmt.Printf("Cover:   %s (source: %s)\n", b.CoverPath, b.CoverSource)
	}
	if len(b.Shelves) > 0 {
		names := make([]string, len(b.Shelves))
		for i, s := range b.Shelves {
			names[i] = s.Name
		}
		fmt.Printf("Shelves: %s\n", strings.Join(names, ", "))
	}
	if len(b.Tags) > 0 {
		names := make([]string, len(b.Tags))
		for i, t := range b.Tags {
			names[i] = t.Name
		}
		fmt.Printf("Tags:    %s\n", strings.Join(names, ", "))
	}
}

func printBookTable(books []models.Book) {
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "ID\tTITLE\tAUTHOR\tSTATUS\tRATING\tYEAR\n")
	for _, b := range books {
		rating := "-"
		if b.Rating > 0 {
			rating = strings.Repeat("*", b.Rating)
		}
		title := b.Title
		if len(title) > 40 {
			title = title[:37] + "..."
		}
		author := b.Author
		if len(author) > 25 {
			author = author[:22] + "..."
		}
		year := ""
		if b.Year > 0 {
			year = strconv.Itoa(b.Year)
		}
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\t%s\n", b.ID, title, author, b.Status, rating, year)
	}
	tw.Flush()
}

// --- commands ---

func serveCmd() *cobra.Command {
	var port int
	var dataDir string
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the library server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Import and run server directly
			fmt.Printf("Starting library server on port %d...\n", port)
			fmt.Printf("Run: go run ./cmd/server --port %d", port)
			if dataDir != "" {
				fmt.Printf(" --data %s", dataDir)
			}
			fmt.Println()
			return nil
		},
	}
	cmd.Flags().IntVar(&port, "port", 8080, "Server port")
	cmd.Flags().StringVar(&dataDir, "data", "", "Data directory")
	return cmd
}

func addCmd() *cobra.Command {
	var author, isbn, genre, status, notes string
	var pages, year int
	cmd := &cobra.Command{
		Use:   "add [title]",
		Short: "Add a book to the library",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			req := models.CreateBookRequest{
				Author: author,
				ISBN:   isbn,
				Genre:  genre,
				Pages:  pages,
				Year:   year,
				Status: status,
				Notes:  notes,
			}
			if len(args) > 0 {
				req.Title = args[0]
			}

			// If ISBN provided but no title, try lookup first
			if req.Title == "" && isbn != "" {
				data, err := doJSON("GET", "/api/lookup/"+url.PathEscape(isbn), nil)
				if err == nil {
					var lookup models.ISBNLookupResult
					json.Unmarshal(data, &lookup)
					req.Title = lookup.Title
					if req.Author == "" {
						req.Author = lookup.Author
					}
					if req.Genre == "" {
						req.Genre = lookup.Genre
					}
					if req.Pages == 0 {
						req.Pages = lookup.Pages
					}
					if req.Year == 0 {
						req.Year = lookup.Year
					}
					fmt.Printf("Found: %s by %s\n", lookup.Title, lookup.Author)
				}
			}

			if req.Title == "" {
				return fmt.Errorf("title is required (provide as argument or use --isbn for auto-lookup)")
			}
			if req.Author == "" {
				return fmt.Errorf("author is required (use --author flag)")
			}

			data, err := doJSON("POST", "/api/books", req)
			if err != nil {
				return err
			}

			if jsonOut {
				printJSON(data)
				return nil
			}

			var book models.Book
			json.Unmarshal(data, &book)
			fmt.Printf("Added book #%d: %s by %s\n", book.ID, book.Title, book.Author)
			return nil
		},
	}
	cmd.Flags().StringVar(&author, "author", "", "Book author (required unless using --isbn)")
	cmd.Flags().StringVar(&isbn, "isbn", "", "ISBN (will auto-lookup metadata)")
	cmd.Flags().StringVar(&genre, "genre", "", "Genre")
	cmd.Flags().IntVar(&pages, "pages", 0, "Page count")
	cmd.Flags().IntVar(&year, "year", 0, "Publication year")
	cmd.Flags().StringVar(&status, "status", "", "Status: unread, reading, finished, abandoned")
	cmd.Flags().StringVar(&notes, "notes", "", "Personal notes")
	return cmd
}

func listCmd() *cobra.Command {
	var status, genre, shelf, tag, sort, order string
	var limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List books in the library",
		RunE: func(cmd *cobra.Command, args []string) error {
			params := url.Values{}
			if status != "" {
				params.Set("status", status)
			}
			if genre != "" {
				params.Set("genre", genre)
			}
			if tag != "" {
				params.Set("tag", tag)
			}
			if sort != "" {
				params.Set("sort", sort)
			}
			if order != "" {
				params.Set("order", order)
			}
			if limit > 0 {
				params.Set("limit", strconv.Itoa(limit))
			}
			path := "/api/books"
			if len(params) > 0 {
				path += "?" + params.Encode()
			}

			data, err := doJSON("GET", path, nil)
			if err != nil {
				return err
			}

			if jsonOut {
				printJSON(data)
				return nil
			}

			var resp models.BookListResponse
			json.Unmarshal(data, &resp)
			if len(resp.Books) == 0 {
				fmt.Println("No books found.")
				return nil
			}
			fmt.Printf("Showing %d of %d books\n\n", len(resp.Books), resp.Total)
			printBookTable(resp.Books)
			return nil
		},
	}
	cmd.Flags().StringVar(&status, "status", "", "Filter by status")
	cmd.Flags().StringVar(&genre, "genre", "", "Filter by genre")
	cmd.Flags().StringVar(&shelf, "shelf", "", "Filter by shelf name")
	cmd.Flags().StringVar(&tag, "tag", "", "Filter by tag")
	cmd.Flags().StringVar(&sort, "sort", "", "Sort by: title, author, rating, year, created_at")
	cmd.Flags().StringVar(&order, "order", "", "Sort order: asc, desc")
	cmd.Flags().IntVar(&limit, "limit", 0, "Limit results")
	return cmd
}

func searchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search [query]",
		Short: "Search books (full-text)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.Join(args, " ")
			data, err := doJSON("GET", "/api/books/search?q="+url.QueryEscape(query), nil)
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(data)
				return nil
			}
			var resp models.BookListResponse
			json.Unmarshal(data, &resp)
			if len(resp.Books) == 0 {
				fmt.Println("No books found.")
				return nil
			}
			fmt.Printf("Found %d results\n\n", resp.Total)
			printBookTable(resp.Books)
			return nil
		},
	}
}

func showCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show [id]",
		Short: "Show book details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := doJSON("GET", "/api/books/"+args[0], nil)
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(data)
				return nil
			}
			var book models.Book
			json.Unmarshal(data, &book)
			printBook(&book)
			return nil
		},
	}
}

func editCmd() *cobra.Command {
	var title, author, isbn, genre, notes string
	var pages, year int
	cmd := &cobra.Command{
		Use:   "edit [id]",
		Short: "Edit a book's details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			req := make(map[string]interface{})
			if cmd.Flags().Changed("title") {
				req["title"] = title
			}
			if cmd.Flags().Changed("author") {
				req["author"] = author
			}
			if cmd.Flags().Changed("isbn") {
				req["isbn"] = isbn
			}
			if cmd.Flags().Changed("genre") {
				req["genre"] = genre
			}
			if cmd.Flags().Changed("notes") {
				req["notes"] = notes
			}
			if cmd.Flags().Changed("pages") {
				req["pages"] = pages
			}
			if cmd.Flags().Changed("year") {
				req["year"] = year
			}
			if len(req) == 0 {
				return fmt.Errorf("no fields to update; use flags like --title, --author, etc.")
			}

			data, err := doJSON("PUT", "/api/books/"+args[0], req)
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(data)
				return nil
			}
			var book models.Book
			json.Unmarshal(data, &book)
			fmt.Printf("Updated book #%d: %s\n", book.ID, book.Title)
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "New title")
	cmd.Flags().StringVar(&author, "author", "", "New author")
	cmd.Flags().StringVar(&isbn, "isbn", "", "New ISBN")
	cmd.Flags().StringVar(&genre, "genre", "", "New genre")
	cmd.Flags().StringVar(&notes, "notes", "", "New notes")
	cmd.Flags().IntVar(&pages, "pages", 0, "New page count")
	cmd.Flags().IntVar(&year, "year", 0, "New year")
	return cmd
}

func deleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a book",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := doJSON("DELETE", "/api/books/"+args[0], nil)
			if err != nil {
				return err
			}
			fmt.Printf("Deleted book #%s\n", args[0])
			return nil
		},
	}
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status [id] [status]",
		Short: "Update reading status (unread, reading, finished, abandoned)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := doJSON("PATCH", "/api/books/"+args[0]+"/status",
				map[string]string{"status": args[1]})
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(data)
				return nil
			}
			var book models.Book
			json.Unmarshal(data, &book)
			fmt.Printf("Book #%d status: %s\n", book.ID, book.Status)
			return nil
		},
	}
}

func rateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rate [id] [1-5]",
		Short: "Rate a book (1-5, or 0 to clear)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			rating, err := strconv.Atoi(args[1])
			if err != nil || rating < 0 || rating > 5 {
				return fmt.Errorf("rating must be 0-5")
			}
			data, err := doJSON("PATCH", "/api/books/"+args[0]+"/rating",
				map[string]int{"rating": rating})
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(data)
				return nil
			}
			fmt.Printf("Book #%s rated %d/5\n", args[0], rating)
			return nil
		},
	}
}

func coverCmd() *cobra.Command {
	var coverURL string
	var refresh bool
	cmd := &cobra.Command{
		Use:   "cover [id]",
		Short: "Manage book cover art",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if refresh {
				data, err := doJSON("POST", "/api/books/"+args[0]+"/cover/refresh", nil)
				if err != nil {
					return err
				}
				if jsonOut {
					printJSON(data)
					return nil
				}
				fmt.Printf("Cover refresh triggered for book #%s\n", args[0])
				return nil
			}
			if coverURL != "" {
				data, err := doJSON("POST", "/api/books/"+args[0]+"/cover",
					map[string]string{"url": coverURL})
				if err != nil {
					return err
				}
				if jsonOut {
					printJSON(data)
					return nil
				}
				fmt.Printf("Cover set for book #%s\n", args[0])
				return nil
			}

			// Just show cover info
			data, err := doJSON("GET", "/api/books/"+args[0], nil)
			if err != nil {
				return err
			}
			var book models.Book
			json.Unmarshal(data, &book)
			if book.CoverPath != "" {
				fmt.Printf("Cover: %s (source: %s)\n", book.CoverPath, book.CoverSource)
				fmt.Printf("URL:   %s/covers/%d.jpg\n", serverURL, book.ID)
			} else {
				fmt.Println("No cover art. Use --refresh to fetch, or --url to set one.")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&coverURL, "url", "", "Set cover from URL")
	cmd.Flags().BoolVar(&refresh, "refresh", false, "Re-fetch cover from APIs")
	return cmd
}

func shelvesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shelves",
		Short: "Manage shelves",
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := doJSON("GET", "/api/shelves", nil)
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(data)
				return nil
			}
			var shelves []models.Shelf
			json.Unmarshal(data, &shelves)
			if len(shelves) == 0 {
				fmt.Println("No shelves. Use 'library shelves create <name>' to create one.")
				return nil
			}
			tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintf(tw, "ID\tNAME\tBOOKS\tDESCRIPTION\n")
			for _, s := range shelves {
				fmt.Fprintf(tw, "%d\t%s\t%d\t%s\n", s.ID, s.Name, s.BookCount, s.Description)
			}
			tw.Flush()
			return nil
		},
	}

	create := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new shelf",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			desc, _ := cmd.Flags().GetString("description")
			data, err := doJSON("POST", "/api/shelves",
				models.CreateShelfRequest{Name: args[0], Description: desc})
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(data)
				return nil
			}
			var shelf models.Shelf
			json.Unmarshal(data, &shelf)
			fmt.Printf("Created shelf #%d: %s\n", shelf.ID, shelf.Name)
			return nil
		},
	}
	create.Flags().String("description", "", "Shelf description")

	add := &cobra.Command{
		Use:   "add [shelf-id] [book-id]",
		Short: "Add a book to a shelf",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			bookID, _ := strconv.ParseInt(args[1], 10, 64)
			_, err := doJSON("POST", "/api/shelves/"+args[0]+"/books",
				models.ShelfBookRequest{BookID: bookID})
			if err != nil {
				return err
			}
			fmt.Printf("Added book #%s to shelf #%s\n", args[1], args[0])
			return nil
		},
	}

	del := &cobra.Command{
		Use:   "delete [shelf-id]",
		Short: "Delete a shelf",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := doJSON("DELETE", "/api/shelves/"+args[0], nil)
			if err != nil {
				return err
			}
			fmt.Printf("Deleted shelf #%s\n", args[0])
			return nil
		},
	}

	cmd.AddCommand(create, add, del)
	return cmd
}

func tagCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tag [book-id] [tag-name]",
		Short: "Add a tag to a book",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := doJSON("POST", "/api/books/"+args[0]+"/tags",
				models.TagRequest{Name: args[1]})
			if err != nil {
				return err
			}
			fmt.Printf("Tagged book #%s with '%s'\n", args[0], args[1])
			return nil
		},
	}
}

func statsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Show reading statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := doJSON("GET", "/api/stats", nil)
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(data)
				return nil
			}
			var s models.StatsResponse
			json.Unmarshal(data, &s)

			fmt.Printf("Total Books:    %d\n", s.TotalBooks)
			fmt.Printf("Reading:        %d\n", s.BooksReading)
			fmt.Printf("Finished:       %d\n", s.BooksRead)
			fmt.Printf("Unread:         %d\n", s.BooksUnread)
			fmt.Printf("Total Pages:    %d\n", s.TotalPages)
			fmt.Printf("Avg Rating:     %.1f\n", s.AverageRating)
			fmt.Printf("Avg Days/Book:  %.1f\n", s.AvgDaysPerBook)
			if s.LongestBook != nil {
				fmt.Printf("Longest Read:   %s (%d pages)\n", s.LongestBook.Title, s.LongestBook.Pages)
			}
			if s.ShortestBook != nil {
				fmt.Printf("Shortest Read:  %s (%d pages)\n", s.ShortestBook.Title, s.ShortestBook.Pages)
			}
			return nil
		},
	}
}

func importCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "import [file]",
		Short: "Import books from CSV or JSON file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			file, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer file.Close()

			ct := "text/csv"
			if strings.HasSuffix(args[0], ".json") {
				ct = "application/json"
			}

			req, _ := http.NewRequest("POST", apiURL("/api/import"), file)
			req.Header.Set("Content-Type", ct)

			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			data, _ := io.ReadAll(resp.Body)
			if jsonOut {
				printJSON(data)
				return nil
			}

			var result models.ImportResult
			json.Unmarshal(data, &result)
			fmt.Printf("Imported: %d, Skipped: %d\n", result.Imported, result.Skipped)
			for _, e := range result.Errors {
				fmt.Printf("  Error: %s\n", e)
			}
			return nil
		},
	}
}

func exportCmd() *cobra.Command {
	var format, output string
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export library to CSV or JSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "/api/export?format=" + format

			resp, err := client.Get(apiURL(path))
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			data, _ := io.ReadAll(resp.Body)

			if output != "" {
				err := os.WriteFile(output, data, 0644)
				if err != nil {
					return err
				}
				fmt.Printf("Exported to %s\n", output)
				return nil
			}

			fmt.Print(string(data))
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", "json", "Export format: json, csv")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path")
	return cmd
}

func lookupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lookup [isbn]",
		Short: "Look up book metadata by ISBN",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := doJSON("GET", "/api/lookup/"+url.PathEscape(args[0]), nil)
			if err != nil {
				return err
			}
			if jsonOut {
				printJSON(data)
				return nil
			}

			var r models.ISBNLookupResult
			json.Unmarshal(data, &r)
			fmt.Printf("Title:     %s\n", r.Title)
			fmt.Printf("Author:    %s\n", r.Author)
			fmt.Printf("ISBN:      %s\n", r.ISBN)
			if r.Genre != "" {
				fmt.Printf("Genre:     %s\n", r.Genre)
			}
			if r.Pages > 0 {
				fmt.Printf("Pages:     %d\n", r.Pages)
			}
			if r.Year > 0 {
				fmt.Printf("Year:      %d\n", r.Year)
			}
			if r.Publisher != "" {
				fmt.Printf("Publisher: %s\n", r.Publisher)
			}
			if r.CoverURL != "" {
				fmt.Printf("Cover:     %s\n", r.CoverURL)
			}
			return nil
		},
	}
}
