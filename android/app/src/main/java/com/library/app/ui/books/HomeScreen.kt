package com.library.app.ui.books

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.navigation.NavHostController
import com.library.app.data.LibraryViewModel
import com.library.app.ui.components.BookCover
import com.library.app.ui.components.StatCard

@Composable
fun HomeScreen(vm: LibraryViewModel, serverUrl: String, navController: NavHostController) {
    val stats by vm.stats.collectAsState()
    val books by vm.books.collectAsState()
    val error by vm.error.collectAsState()

    LaunchedEffect(serverUrl) {
        vm.loadStats()
        vm.loadBooks()
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState())
            .padding(16.dp)
    ) {
        Text("My Library", style = MaterialTheme.typography.headlineMedium, fontWeight = FontWeight.Bold)
        Spacer(Modifier.height(16.dp))

        error?.let {
            Card(colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.errorContainer)) {
                Text(it, modifier = Modifier.padding(12.dp), color = MaterialTheme.colorScheme.onErrorContainer)
            }
            Spacer(Modifier.height(12.dp))
        }

        // Stats cards
        stats?.let { s ->
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.fillMaxWidth()) {
                StatCard("Total", "${s.totalBooks}", Modifier.weight(1f))
                StatCard("Reading", "${s.booksReading}", Modifier.weight(1f))
                StatCard("Finished", "${s.booksRead}", Modifier.weight(1f))
                StatCard("Pages", "${s.totalPages}", Modifier.weight(1f))
            }
            Spacer(Modifier.height(20.dp))

            if (s.overdueCount > 0) {
                Card(colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.errorContainer)) {
                    Row(
                        modifier = Modifier.fillMaxWidth().padding(12.dp),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Text("${s.overdueCount} overdue loan(s)", fontWeight = FontWeight.Bold,
                            color = MaterialTheme.colorScheme.onErrorContainer)
                        TextButton(onClick = { navController.navigate("loans") }) { Text("View") }
                    }
                }
                Spacer(Modifier.height(20.dp))
            }
        }

        // Currently reading
        val reading = books.filter { it.status == "reading" }
        if (reading.isNotEmpty()) {
            SectionHeader("Currently Reading") { navController.navigate("books") }
            LazyRow(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                items(reading.take(10)) { book ->
                    BookCover(book = book, serverUrl = serverUrl, onClick = { navController.navigate("books/${book.id}") })
                }
            }
            Spacer(Modifier.height(20.dp))
        }

        // Recent
        val recent = books.sortedByDescending { it.createdAt }.take(10)
        if (recent.isNotEmpty()) {
            SectionHeader("Recently Added") { navController.navigate("books") }
            LazyRow(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                items(recent) { book ->
                    BookCover(book = book, serverUrl = serverUrl, onClick = { navController.navigate("books/${book.id}") })
                }
            }
        }

        if (books.isEmpty() && stats?.totalBooks == 0) {
            Spacer(Modifier.height(40.dp))
            Column(horizontalAlignment = Alignment.CenterHorizontally, modifier = Modifier.fillMaxWidth()) {
                Text("Your library is empty", style = MaterialTheme.typography.titleMedium)
                Spacer(Modifier.height(8.dp))
                Button(onClick = { navController.navigate("books/add") }) { Text("Add your first book") }
            }
        }
    }
}

@Composable
fun SectionHeader(title: String, onViewAll: () -> Unit) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(title, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.SemiBold)
        TextButton(onClick = onViewAll) { Text("View all") }
    }
}
