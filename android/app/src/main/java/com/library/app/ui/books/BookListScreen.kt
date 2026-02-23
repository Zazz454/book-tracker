package com.library.app.ui.books

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Search
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.navigation.NavHostController
import com.library.app.data.LibraryViewModel
import com.library.app.ui.components.BookCover

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun BookListScreen(vm: LibraryViewModel, serverUrl: String, navController: NavHostController) {
    val books by vm.books.collectAsState()
    val isLoading by vm.isLoading.collectAsState()
    val searchQuery by vm.searchQuery.collectAsState()
    var statusFilter by remember { mutableStateOf<String?>(null) }

    LaunchedEffect(statusFilter, serverUrl) {
        vm.loadBooks(status = statusFilter)
    }

    Column(modifier = Modifier.fillMaxSize()) {
        // Search bar
        OutlinedTextField(
            value = searchQuery,
            onValueChange = {
                vm.setSearchQuery(it)
                if (it.length >= 2) vm.loadBooks(query = it)
                else if (it.isEmpty()) vm.loadBooks(status = statusFilter)
            },
            placeholder = { Text("Search books...") },
            leadingIcon = { Icon(Icons.Default.Search, contentDescription = null) },
            modifier = Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 8.dp),
            singleLine = true,
        )

        // Status filter chips
        Row(
            modifier = Modifier.fillMaxWidth().padding(horizontal = 16.dp),
            horizontalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            listOf(null to "All", "unread" to "Unread", "reading" to "Reading", "finished" to "Finished").forEach { (s, label) ->
                FilterChip(
                    selected = statusFilter == s,
                    onClick = { statusFilter = s; vm.setSearchQuery(""); vm.loadBooks(status = s) },
                    label = { Text(label) },
                )
            }
        }

        if (isLoading) {
            LinearProgressIndicator(modifier = Modifier.fillMaxWidth())
        }

        // Book grid
        LazyVerticalGrid(
            columns = GridCells.Adaptive(110.dp),
            contentPadding = PaddingValues(16.dp),
            horizontalArrangement = Arrangement.spacedBy(10.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            items(books) { book ->
                BookCover(
                    book = book,
                    serverUrl = serverUrl,
                    onClick = { navController.navigate("books/${book.id}") },
                )
            }
        }
    }
}
