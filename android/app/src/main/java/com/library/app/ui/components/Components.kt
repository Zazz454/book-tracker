package com.library.app.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import coil.compose.AsyncImage
import com.library.app.api.ApiClient
import com.library.app.data.Book

@Composable
fun BookCover(book: Book, serverUrl: String, modifier: Modifier = Modifier, onClick: () -> Unit = {}) {
    Column(
        modifier = modifier
            .width(110.dp)
            .clickable(onClick = onClick),
    ) {
        Card(
            modifier = Modifier.fillMaxWidth().aspectRatio(2f / 3f),
            shape = RoundedCornerShape(8.dp),
        ) {
            if (book.coverPath.isNotEmpty()) {
                AsyncImage(
                    model = ApiClient.coverUrl(serverUrl, book.coverPath),
                    contentDescription = book.title,
                    contentScale = ContentScale.Crop,
                    modifier = Modifier.fillMaxSize(),
                )
            } else {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .background(MaterialTheme.colorScheme.surfaceVariant),
                    contentAlignment = Alignment.Center,
                ) {
                    Text(
                        book.title.take(30),
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                        modifier = Modifier.padding(8.dp),
                    )
                }
            }
        }
        Spacer(Modifier.height(4.dp))
        Text(
            book.title,
            style = MaterialTheme.typography.bodySmall,
            fontWeight = FontWeight.Medium,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
        )
        Text(
            book.author,
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
            fontSize = 11.sp,
        )
    }
}

@Composable
fun StatCard(label: String, value: String, modifier: Modifier = Modifier) {
    Card(modifier = modifier) {
        Column(
            modifier = Modifier.padding(12.dp).fillMaxWidth(),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Text(value, style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.Bold,
                color = MaterialTheme.colorScheme.primary)
            Text(label, style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant)
        }
    }
}

@Composable
fun StatusBadge(status: String, modifier: Modifier = Modifier) {
    val (color, label) = when (status) {
        "reading" -> Color(0xFF0071E3) to "Reading"
        "finished" -> Color(0xFF34C759) to "Finished"
        "abandoned" -> Color(0xFFFF3B30) to "Abandoned"
        else -> Color(0xFF8E8E93) to "Unread"
    }
    Surface(
        color = color,
        shape = RoundedCornerShape(4.dp),
        modifier = modifier,
    ) {
        Text(
            label,
            color = Color.White,
            fontSize = 11.sp,
            fontWeight = FontWeight.SemiBold,
            modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp),
        )
    }
}

@Composable
fun RatingStars(rating: Int, onRate: (Int) -> Unit = {}, interactive: Boolean = true) {
    Row {
        (1..5).forEach { star ->
            val text = if (star <= rating) "\u2605" else "\u2606"
            Text(
                text,
                fontSize = 24.sp,
                color = if (star <= rating) Color(0xFFFF9500) else MaterialTheme.colorScheme.outline,
                modifier = if (interactive) Modifier.clickable { onRate(star) } else Modifier,
            )
        }
    }
}
