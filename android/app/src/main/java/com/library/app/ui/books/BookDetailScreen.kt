package com.library.app.ui.books

import android.content.Intent
import android.net.Uri
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.navigation.NavHostController
import coil.compose.AsyncImage
import com.library.app.api.ApiClient
import com.library.app.data.LibraryViewModel
import com.library.app.ui.components.RatingStars
import com.library.app.ui.components.StatusBadge
import java.net.URLEncoder

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun BookDetailScreen(vm: LibraryViewModel, bookId: Long, serverUrl: String, navController: NavHostController) {
    val book by vm.selectedBook.collectAsState()
    val loans by vm.bookLoans.collectAsState()
    val context = LocalContext.current

    LaunchedEffect(bookId) { vm.loadBook(bookId) }

    val b = book ?: return Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
        CircularProgressIndicator()
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Book Detail") },
                navigationIcon = { IconButton(onClick = { navController.popBackStack() }) {
                    Icon(Icons.Default.ArrowBack, "Back")
                }},
                actions = {
                    IconButton(onClick = {
                        vm.deleteBook(b.id) { navController.popBackStack() }
                    }) { Icon(Icons.Default.Delete, "Delete", tint = MaterialTheme.colorScheme.error) }
                }
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier.padding(padding).fillMaxSize().verticalScroll(rememberScrollState()).padding(16.dp)
        ) {
            // Cover
            if (b.coverPath.isNotEmpty()) {
                Card(modifier = Modifier.fillMaxWidth().aspectRatio(2f / 3f).heightIn(max = 300.dp)) {
                    AsyncImage(
                        model = ApiClient.coverUrl(serverUrl, b.coverPath),
                        contentDescription = b.title,
                        contentScale = ContentScale.Fit,
                        modifier = Modifier.fillMaxSize(),
                    )
                }
                Spacer(Modifier.height(12.dp))
            }

            Text(b.title, style = MaterialTheme.typography.headlineSmall, fontWeight = FontWeight.Bold)
            Text("by ${b.author}", style = MaterialTheme.typography.titleMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant)
            Spacer(Modifier.height(12.dp))

            // Status + actions
            Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                StatusBadge(b.status)
                if (b.status != "reading") TextButton(onClick = { vm.updateStatus(b.id, "reading") }) { Text("Start") }
                if (b.status != "finished") TextButton(onClick = { vm.updateStatus(b.id, "finished") }) { Text("Finish") }
            }
            Spacer(Modifier.height(8.dp))

            // Rating
            RatingStars(b.rating, onRate = { vm.updateRating(b.id, it) })
            Spacer(Modifier.height(12.dp))

            // Fields
            if (b.isbn.isNotEmpty()) DetailField("ISBN", b.isbn)
            if (b.genre.isNotEmpty()) DetailField("Genre", b.genre)
            if (b.pages > 0) DetailField("Pages", "${b.pages}")
            if (b.year > 0) DetailField("Year", "${b.year}")
            Spacer(Modifier.height(12.dp))

            // Active loan
            b.activeLoan?.let { loan ->
                Card(colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.secondaryContainer)) {
                    Column(Modifier.padding(12.dp)) {
                        val label = if (loan.loanType == "lent") "Lent to" else "Borrowed from"
                        Text("$label ${loan.personName}", fontWeight = FontWeight.SemiBold)
                        if (loan.isOverdue) Text("${loan.daysOverdue} days overdue",
                            color = MaterialTheme.colorScheme.error, fontWeight = FontWeight.Bold)
                        Spacer(Modifier.height(8.dp))
                        Button(onClick = { vm.checkInLoan(loan.id) { vm.loadBook(bookId) } }) { Text("Return") }
                    }
                }
                Spacer(Modifier.height(12.dp))
            } ?: run {
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    OutlinedButton(onClick = { navController.navigate("loans/add?bookId=${b.id}&type=lent") }) { Text("Lend") }
                    OutlinedButton(onClick = { navController.navigate("loans/add?bookId=${b.id}&type=borrowed") }) { Text("Borrow") }
                }
                Spacer(Modifier.height(12.dp))
            }

            // External links
            Text("Find Online", fontWeight = FontWeight.SemiBold)
            Spacer(Modifier.height(6.dp))
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                val q = URLEncoder.encode("${b.title} ${b.author}", "UTF-8")
                FilledTonalButton(onClick = {
                    context.startActivity(Intent(Intent.ACTION_VIEW, Uri.parse("https://www.amazon.com/s?k=$q&i=stripbooks")))
                }) { Text("Amazon") }
                FilledTonalButton(onClick = {
                    context.startActivity(Intent(Intent.ACTION_VIEW, Uri.parse("https://www.worldcat.org/search?q=$q")))
                }) { Text("WorldCat") }
                FilledTonalButton(onClick = {
                    context.startActivity(Intent(Intent.ACTION_VIEW, Uri.parse("https://openlibrary.org/search?q=$q")))
                }) { Text("Open Library") }
            }
            Spacer(Modifier.height(16.dp))

            // Notes
            if (b.notes.isNotEmpty()) {
                Text("Notes", fontWeight = FontWeight.SemiBold)
                Text(b.notes, color = MaterialTheme.colorScheme.onSurfaceVariant)
                Spacer(Modifier.height(12.dp))
            }

            // Loan history
            if (loans.isNotEmpty()) {
                Text("Loan History", fontWeight = FontWeight.SemiBold)
                Spacer(Modifier.height(6.dp))
                loans.forEach { loan ->
                    Card(modifier = Modifier.fillMaxWidth().padding(vertical = 2.dp)) {
                        Row(modifier = Modifier.padding(10.dp).fillMaxWidth(),
                            horizontalArrangement = Arrangement.SpaceBetween) {
                            Column {
                                val label = if (loan.loanType == "lent") "Lent to" else "Borrowed from"
                                Text("$label ${loan.personName}", style = MaterialTheme.typography.bodyMedium)
                                Text(loan.checkedOut.take(10), style = MaterialTheme.typography.bodySmall,
                                    color = MaterialTheme.colorScheme.onSurfaceVariant)
                            }
                            if (loan.checkedIn != null) Text("Returned", color = MaterialTheme.colorScheme.primary)
                            else if (loan.isOverdue) Text("Overdue", color = MaterialTheme.colorScheme.error, fontWeight = FontWeight.Bold)
                            else Text("Active", color = MaterialTheme.colorScheme.tertiary)
                        }
                    }
                }
            }

            // Refresh cover
            Spacer(Modifier.height(16.dp))
            OutlinedButton(onClick = { vm.refreshCover(b.id) }, modifier = Modifier.fillMaxWidth()) {
                Text("Refresh Cover")
            }
        }
    }
}

@Composable
fun DetailField(label: String, value: String) {
    Row(modifier = Modifier.fillMaxWidth().padding(vertical = 4.dp)) {
        Text(label, fontWeight = FontWeight.Medium, color = MaterialTheme.colorScheme.onSurfaceVariant,
            modifier = Modifier.width(80.dp))
        Text(value)
    }
}
