package com.library.app.ui

import android.os.Build
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext

private val DarkColorScheme = darkColorScheme(
    primary = Color(0xFF4DA6FF),
    onPrimary = Color.White,
    secondary = Color(0xFF80CAFF),
    background = Color(0xFF1A1A2E),
    surface = Color(0xFF1E1E3A),
    surfaceVariant = Color(0xFF252547),
    onBackground = Color(0xFFF0F0F5),
    onSurface = Color(0xFFF0F0F5),
    error = Color(0xFFFF3B30),
    outline = Color(0xFF2E2E4A),
)

private val LightColorScheme = lightColorScheme(
    primary = Color(0xFF0071E3),
    onPrimary = Color.White,
    secondary = Color(0xFF005BB5),
    background = Color(0xFFF5F5F7),
    surface = Color.White,
    surfaceVariant = Color(0xFFF0F0F2),
    onBackground = Color(0xFF1D1D1F),
    onSurface = Color(0xFF1D1D1F),
    error = Color(0xFFFF3B30),
    outline = Color(0xFFD2D2D7),
)

@Composable
fun LibraryTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),
    content: @Composable () -> Unit,
) {
    val colorScheme = when {
        Build.VERSION.SDK_INT >= Build.VERSION_CODES.S -> {
            val context = LocalContext.current
            if (darkTheme) dynamicDarkColorScheme(context) else dynamicLightColorScheme(context)
        }
        darkTheme -> DarkColorScheme
        else -> LightColorScheme
    }

    MaterialTheme(
        colorScheme = colorScheme,
        content = content,
    )
}
