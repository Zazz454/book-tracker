package com.library.app.ui.loans

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.navigation.NavHostController
import com.library.app.data.LibraryViewModel
import com.library.app.data.Loan

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun LoanListScreen(vm: LibraryViewModel, navController: NavHostController) {
    val loans by vm.loans.collectAsState()
    val isLoading by vm.isLoading.collectAsState()
    var selectedTab by remember { mutableStateOf(0) }
    val tabs = listOf("active", "overdue", "returned")
    val tabLabels = listOf("Active", "Overdue", "Returned")

    LaunchedEffect(selectedTab) { vm.loadLoans(tabs[selectedTab]) }

    Scaffold(
        topBar = { TopAppBar(title = { Text("Loans") }) },
        floatingActionButton = {
            FloatingActionButton(onClick = { navController.navigate("loans/add?bookId=0&type=lent") }) {
                Icon(Icons.Default.Add, "New Loan")
            }
        }
    ) { padding ->
        Column(modifier = Modifier.padding(padding)) {
            TabRow(selectedTabIndex = selectedTab) {
                tabLabels.forEachIndexed { index, title ->
                    Tab(selected = selectedTab == index, onClick = { selectedTab = index },
                        text = { Text(title) })
                }
            }

            if (isLoading) LinearProgressIndicator(modifier = Modifier.fillMaxWidth())

            if (loans.isEmpty() && !isLoading) {
                Box(modifier = Modifier.fillMaxSize().padding(32.dp)) {
                    Text("No ${tabLabels[selectedTab].lowercase()} loans", color = MaterialTheme.colorScheme.onSurfaceVariant)
                }
            } else {
                LazyColumn(contentPadding = PaddingValues(16.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                    items(loans) { loan ->
                        LoanCard(loan = loan, onCheckIn = {
                            vm.checkInLoan(loan.id) { vm.loadLoans(tabs[selectedTab]) }
                        }, onTap = {
                            loan.book?.let { navController.navigate("books/${it.id}") }
                        })
                    }
                }
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun LoanCard(loan: Loan, onCheckIn: () -> Unit, onTap: () -> Unit) {
    Card(
        onClick = onTap,
        modifier = Modifier.fillMaxWidth(),
    ) {
        Column(modifier = Modifier.padding(12.dp)) {
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                Column(modifier = Modifier.weight(1f)) {
                    Text(loan.book?.title ?: "Book #${loan.bookId}", fontWeight = FontWeight.SemiBold)
                    Text(loan.book?.author ?: "", style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant)
                }
                Surface(
                    color = if (loan.loanType == "lent") MaterialTheme.colorScheme.primary else Color(0xFFFF9500),
                    shape = MaterialTheme.shapes.small,
                ) {
                    Text(
                        if (loan.loanType == "lent") "Lent" else "Borrowed",
                        color = Color.White,
                        style = MaterialTheme.typography.labelSmall,
                        modifier = Modifier.padding(horizontal = 8.dp, vertical = 2.dp),
                    )
                }
            }
            Spacer(Modifier.height(6.dp))
            val prefix = if (loan.loanType == "lent") "To" else "From"
            Text("$prefix: ${loan.personName}", style = MaterialTheme.typography.bodyMedium)
            loan.dueDate?.let { due ->
                Text("Due: ${due.take(10)}",
                    color = if (loan.isOverdue) MaterialTheme.colorScheme.error else MaterialTheme.colorScheme.onSurfaceVariant,
                    fontWeight = if (loan.isOverdue) FontWeight.Bold else FontWeight.Normal,
                    style = MaterialTheme.typography.bodySmall)
            }
            if (loan.isOverdue) {
                Text("${loan.daysOverdue} days overdue", color = MaterialTheme.colorScheme.error,
                    fontWeight = FontWeight.Bold, style = MaterialTheme.typography.bodySmall)
            }
            loan.checkedIn?.let {
                Text("Returned: ${it.take(10)}", color = MaterialTheme.colorScheme.primary,
                    style = MaterialTheme.typography.bodySmall)
            }
            if (loan.checkedIn == null) {
                Spacer(Modifier.height(8.dp))
                Button(onClick = onCheckIn, modifier = Modifier.fillMaxWidth()) { Text("Return") }
            }
        }
    }
}
