package com.library.app.ui.loans

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.navigation.NavHostController
import com.library.app.data.CreateLoanRequest
import com.library.app.data.LibraryViewModel
import java.time.LocalDate
import java.time.format.DateTimeFormatter

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AddLoanScreen(vm: LibraryViewModel, bookId: Long?, loanType: String, navController: NavHostController) {
    var personName by remember { mutableStateOf("") }
    var personContact by remember { mutableStateOf("") }
    var dueDate by remember { mutableStateOf(LocalDate.now().plusWeeks(2).format(DateTimeFormatter.ISO_LOCAL_DATE)) }
    var notes by remember { mutableStateOf("") }
    var selectedType by remember { mutableStateOf(loanType) }
    var bookIdInput by remember { mutableStateOf(if (bookId != null && bookId > 0) "$bookId" else "") }
    val isLoading by vm.isLoading.collectAsState()
    val error by vm.error.collectAsState()

    val title = if (selectedType == "borrowed") "Borrow a Book" else "Lend a Book"

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(title) },
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

            // Loan type toggle
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                FilterChip(selected = selectedType == "lent", onClick = { selectedType = "lent" },
                    label = { Text("Lend") })
                FilterChip(selected = selectedType == "borrowed", onClick = { selectedType = "borrowed" },
                    label = { Text("Borrow") })
            }
            Spacer(Modifier.height(12.dp))

            if (bookId == null || bookId <= 0) {
                OutlinedTextField(value = bookIdInput, onValueChange = { bookIdInput = it },
                    label = { Text("Book ID") }, modifier = Modifier.fillMaxWidth(), singleLine = true)
                Spacer(Modifier.height(8.dp))
            }

            val personLabel = if (selectedType == "borrowed") "Borrowed From" else "Lent To"
            OutlinedTextField(value = personName, onValueChange = { personName = it },
                label = { Text("$personLabel *") }, modifier = Modifier.fillMaxWidth(), singleLine = true)
            Spacer(Modifier.height(8.dp))

            OutlinedTextField(value = personContact, onValueChange = { personContact = it },
                label = { Text("Contact (optional)") }, modifier = Modifier.fillMaxWidth(), singleLine = true)
            Spacer(Modifier.height(8.dp))

            OutlinedTextField(value = dueDate, onValueChange = { dueDate = it },
                label = { Text("Due Date (YYYY-MM-DD)") }, modifier = Modifier.fillMaxWidth(), singleLine = true)
            Spacer(Modifier.height(8.dp))

            OutlinedTextField(value = notes, onValueChange = { notes = it },
                label = { Text("Notes") }, modifier = Modifier.fillMaxWidth(), minLines = 2)
            Spacer(Modifier.height(16.dp))

            val actualBookId = bookId?.takeIf { it > 0 } ?: bookIdInput.toLongOrNull() ?: 0

            Button(
                onClick = {
                    vm.clearError()
                    vm.createLoan(
                        CreateLoanRequest(
                            bookId = actualBookId,
                            loanType = selectedType,
                            personName = personName.trim(),
                            personContact = personContact.trim(),
                            dueDate = dueDate.trim(),
                            notes = notes.trim(),
                        )
                    ) { navController.popBackStack() }
                },
                modifier = Modifier.fillMaxWidth(),
                enabled = personName.isNotBlank() && actualBookId > 0 && !isLoading,
            ) {
                if (isLoading) CircularProgressIndicator(modifier = Modifier.size(20.dp), strokeWidth = 2.dp)
                else Text(if (selectedType == "borrowed") "Record Borrow" else "Lend Book")
            }
        }
    }
}
