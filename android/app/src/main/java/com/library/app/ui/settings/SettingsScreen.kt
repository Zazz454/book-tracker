package com.library.app.ui.settings

import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.library.app.data.LibraryViewModel

@Composable
fun SettingsScreen(vm: LibraryViewModel) {
    val serverUrl by vm.serverUrl.collectAsState()
    var urlInput by remember(serverUrl) { mutableStateOf(serverUrl) }
    var saved by remember { mutableStateOf(false) }

    Column(modifier = Modifier.fillMaxSize().padding(16.dp)) {
        Text("Settings", style = MaterialTheme.typography.headlineMedium, fontWeight = FontWeight.Bold)
        Spacer(Modifier.height(24.dp))

        Text("Server URL", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.SemiBold)
        Spacer(Modifier.height(4.dp))
        Text("Point this to your library server", color = MaterialTheme.colorScheme.onSurfaceVariant,
            style = MaterialTheme.typography.bodySmall)
        Spacer(Modifier.height(8.dp))

        OutlinedTextField(
            value = urlInput,
            onValueChange = { urlInput = it; saved = false },
            label = { Text("Server URL") },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true,
        )
        Spacer(Modifier.height(12.dp))

        Button(
            onClick = { vm.setServerUrl(urlInput.trim()); saved = true },
            modifier = Modifier.fillMaxWidth(),
        ) { Text("Save") }

        if (saved) {
            Spacer(Modifier.height(8.dp))
            Text("Saved. Restart the app or switch tabs to reconnect.",
                color = MaterialTheme.colorScheme.primary, style = MaterialTheme.typography.bodySmall)
        }

        Spacer(Modifier.height(32.dp))
        Divider()
        Spacer(Modifier.height(16.dp))
        Text("About", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.SemiBold)
        Spacer(Modifier.height(8.dp))
        Text("My Library v1.0", style = MaterialTheme.typography.bodyMedium)
        Text("Personal library catalog with lending, covers, and reading stats.",
            color = MaterialTheme.colorScheme.onSurfaceVariant, style = MaterialTheme.typography.bodySmall)
    }
}
