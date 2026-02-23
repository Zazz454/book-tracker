package com.library.app.api

import com.jakewharton.retrofit2.converter.kotlinx.serialization.asConverterFactory
import kotlinx.serialization.json.Json
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import java.util.concurrent.TimeUnit

object ApiClient {
    private val json = Json {
        ignoreUnknownKeys = true
        coerceInputValues = true
        isLenient = true
    }

    private var _api: LibraryApi? = null
    private var _baseUrl: String = ""

    fun getApi(baseUrl: String): LibraryApi {
        if (_api == null || _baseUrl != baseUrl) {
            _baseUrl = baseUrl
            val client = OkHttpClient.Builder()
                .connectTimeout(10, TimeUnit.SECONDS)
                .readTimeout(30, TimeUnit.SECONDS)
                .addInterceptor(
                    HttpLoggingInterceptor().apply {
                        level = HttpLoggingInterceptor.Level.BASIC
                    }
                )
                .build()

            val retrofit = Retrofit.Builder()
                .baseUrl(baseUrl.trimEnd('/') + "/")
                .client(client)
                .addConverterFactory(json.asConverterFactory("application/json".toMediaType()))
                .build()

            _api = retrofit.create(LibraryApi::class.java)
        }
        return _api!!
    }

    fun coverUrl(baseUrl: String, coverPath: String): String {
        if (coverPath.isEmpty()) return ""
        return baseUrl.trimEnd('/') + "/" + coverPath.trimStart('/')
    }
}
