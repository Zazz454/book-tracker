package com.library.app.ui

import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.navigation.NavHostController
import androidx.navigation.compose.*
import com.library.app.data.LibraryViewModel
import com.library.app.ui.books.*
import com.library.app.ui.loans.*
import com.library.app.ui.settings.SettingsScreen

sealed class Screen(val route: String, val title: String) {
    data object Home : Screen("home", "Home")
    data object Books : Screen("books", "Books")
    data object BookDetail : Screen("books/{id}", "Book")
    data object AddBook : Screen("books/add", "Add Book")
    data object Loans : Screen("loans", "Loans")
    data object AddLoan : Screen("loans/add?bookId={bookId}&type={type}", "New Loan")
    data object Settings : Screen("settings", "Settings")
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun LibraryNavHost(vm: LibraryViewModel = viewModel()) {
    val navController = rememberNavController()
    val currentRoute = navController.currentBackStackEntryAsState().value?.destination?.route

    val serverUrl by vm.serverUrl.collectAsState()

    Scaffold(
        bottomBar = {
            NavigationBar {
                NavigationBarItem(
                    selected = currentRoute == Screen.Home.route,
                    onClick = { navController.navigateSingle(Screen.Home.route) },
                    icon = { Icon(Icons.Default.Home, contentDescription = "Home") },
                    label = { Text("Home") },
                )
                NavigationBarItem(
                    selected = currentRoute == Screen.Books.route,
                    onClick = { navController.navigateSingle(Screen.Books.route) },
                    icon = { Icon(Icons.Default.Book, contentDescription = "Books") },
                    label = { Text("Books") },
                )
                NavigationBarItem(
                    selected = false,
                    onClick = { navController.navigate("books/add") },
                    icon = { Icon(Icons.Default.Add, contentDescription = "Add") },
                    label = { Text("Add") },
                )
                NavigationBarItem(
                    selected = currentRoute == Screen.Loans.route,
                    onClick = { navController.navigateSingle(Screen.Loans.route) },
                    icon = {
                        val stats by vm.stats.collectAsState()
                        BadgedBox(
                            badge = {
                                val count = stats?.overdueCount ?: 0
                                if (count > 0) Badge { Text("$count") }
                            }
                        ) { Icon(Icons.Default.SwapHoriz, contentDescription = "Loans") }
                    },
                    label = { Text("Loans") },
                )
                NavigationBarItem(
                    selected = currentRoute == Screen.Settings.route,
                    onClick = { navController.navigateSingle(Screen.Settings.route) },
                    icon = { Icon(Icons.Default.Settings, contentDescription = "Settings") },
                    label = { Text("More") },
                )
            }
        }
    ) { padding ->
        NavHost(
            navController = navController,
            startDestination = Screen.Home.route,
            modifier = Modifier.padding(padding),
        ) {
            composable(Screen.Home.route) {
                HomeScreen(vm = vm, serverUrl = serverUrl, navController = navController)
            }
            composable(Screen.Books.route) {
                BookListScreen(vm = vm, serverUrl = serverUrl, navController = navController)
            }
            composable("books/{id}") { backStack ->
                val id = backStack.arguments?.getString("id")?.toLongOrNull() ?: return@composable
                BookDetailScreen(vm = vm, bookId = id, serverUrl = serverUrl, navController = navController)
            }
            composable("books/add") {
                AddBookScreen(vm = vm, navController = navController)
            }
            composable(Screen.Loans.route) {
                LoanListScreen(vm = vm, navController = navController)
            }
            composable("loans/add?bookId={bookId}&type={type}") { backStack ->
                val bookId = backStack.arguments?.getString("bookId")?.toLongOrNull()
                val type = backStack.arguments?.getString("type") ?: "lent"
                AddLoanScreen(vm = vm, bookId = bookId, loanType = type, navController = navController)
            }
            composable(Screen.Settings.route) {
                SettingsScreen(vm = vm)
            }
        }
    }

    // Load stats for badge on first composition
    LaunchedEffect(serverUrl) {
        vm.loadStats()
    }
}

fun NavHostController.navigateSingle(route: String) {
    navigate(route) {
        popUpTo(graph.startDestinationId) { saveState = true }
        launchSingleTop = true
        restoreState = true
    }
}
