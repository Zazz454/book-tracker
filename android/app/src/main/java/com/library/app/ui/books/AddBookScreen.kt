package com.library.app.ui.books

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.navigation.NavHostController
import com.library.app.data.CreateBookRequest
import com.library.app.data.LibraryViewModel

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AddBookScreen(vm: LibraryViewModel, navController: NavHostController) {
    var title by remember { mutableStateOf("") }
    var author by remember { mutableStateOf("") }
    var isbn by remember { mutableStateOf("") }
    var genre by remember { mutableStateOf("") }
    var pages by remember { mutableStateOf("") }
    var year by remember { mutableStateOf("") }
    var notes by remember { mutableStateOf("") }
    val isLoading by vm.isLoading.collectAsState()
    val error by vm.error.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Add Book") },
                navigationIcon = { IconButton(onClick = { navController.popBackStack() }) {
                    Icon(Icons.Default.ArrowBack, "Back")
                }},
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier.padding(padding).fillMaxSize().verticalScroll(rememberScrollState()).padding(16.dp)
        ) {
            error?.let {
                Card(colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.errorContainer)) {
                    Text(it, modifier = Modifier.padding(12.dp), color = MaterialTheme.colorScheme.onErrorContainer)
                }
                Spacer(Modifier.height(12.dp))
            }

            OutlinedTextField(value = title, onValueChange = { title = it }, label = { Text("Title *") },
                modifier = Modifier.fillMaxWidth(), singleLine = true)
            Spacer(Modifier.height(8.dp))

            OutlinedTextField(value = author, onValueChange = { author = it }, label = { Text("Author *") },
                modifier = Modifier.fillMaxWidth(), singleLine = true)
            Spacer(Modifier.height(8.dp))

            OutlinedTextField(value = isbn, onValueChange = { isbn = it }, label = { Text("ISBN") },
                modifier = Modifier.fillMaxWidth(), singleLine = true)
            Spacer(Modifier.height(8.dp))

            OutlinedTextField(value = genre, onValueChange = { genre = it }, label = { Text("Genre") },
                modifier = Modifier.fillMaxWidth(), singleLine = true)
            Spacer(Modifier.height(8.dp))

            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                OutlinedTextField(value = pages, onValueChange = { pages = it }, label = { Text("Pages") },
                    modifier = Modifier.weight(1f), singleLine = true,
                    keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number))
                OutlinedTextField(value = year, onValueChange = { year = it }, label = { Text("Year") },
                    modifier = Modifier.weight(1f), singleLine = true,
                    keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number))
            }
            Spacer(Modifier.height(8.dp))

            OutlinedTextField(value = notes, onValueChange = { notes = it }, label = { Text("Notes") },
                modifier = Modifier.fillMaxWidth(), minLines = 3)
            Spacer(Modifier.height(16.dp))

            Button(
                onClick = {
                    vm.clearError()
                    vm.createBook(
                        CreateBookRequest(
                            title = title.trim(),
                            author = author.trim(),
                            isbn = isbn.trim(),
                            genre = genre.trim(),
                            pages = pages.toIntOrNull() ?: 0,
                            year = year.toIntOrNull() ?: 0,
                            notes = notes.trim(),
                        )
                    ) { navController.popBackStack() }
                },
                modifier = Modifier.fillMaxWidth(),
                enabled = title.isNotBlank() && author.isNotBlank() && !isLoading,
            ) {
                if (isLoading) CircularProgressIndicator(modifier = Modifier.size(20.dp), strokeWidth = 2.dp)
                else Text("Add Book")
            }
        }
    }
}
