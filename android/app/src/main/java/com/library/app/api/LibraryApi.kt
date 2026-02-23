package com.library.app.api

import com.library.app.data.*
import retrofit2.http.*

interface LibraryApi {

    // Books
    @GET("/api/books")
    suspend fun listBooks(
        @Query("status") status: String? = null,
        @Query("genre") genre: String? = null,
        @Query("sort") sort: String? = null,
        @Query("order") order: String? = null,
        @Query("q") query: String? = null,
        @Query("limit") limit: Int = 50,
        @Query("offset") offset: Int = 0,
    ): BookListResponse

    @GET("/api/books/search")
    suspend fun searchBooks(@Query("q") query: String): BookListResponse

    @POST("/api/books")
    suspend fun createBook(@Body request: CreateBookRequest): Book

    @GET("/api/books/{id}")
    suspend fun getBook(@Path("id") id: Long): Book

    @PUT("/api/books/{id}")
    suspend fun updateBook(@Path("id") id: Long, @Body request: Map<String, @JvmSuppressWildcards Any>): Book

    @DELETE("/api/books/{id}")
    suspend fun deleteBook(@Path("id") id: Long)

    @PATCH("/api/books/{id}/status")
    suspend fun updateStatus(@Path("id") id: Long, @Body request: StatusRequest): Book

    @PATCH("/api/books/{id}/rating")
    suspend fun updateRating(@Path("id") id: Long, @Body request: RatingRequest): Book

    @POST("/api/books/{id}/cover/refresh")
    suspend fun refreshCover(@Path("id") id: Long)

    // Loans
    @GET("/api/loans")
    suspend fun listLoans(
        @Query("status") status: String = "active",
        @Query("loan_type") loanType: String? = null,
        @Query("limit") limit: Int = 50,
    ): LoanListResponse

    @POST("/api/loans")
    suspend fun createLoan(@Body request: CreateLoanRequest): Loan

    @GET("/api/loans/{id}")
    suspend fun getLoan(@Path("id") id: Long): Loan

    @PATCH("/api/loans/{id}")
    suspend fun checkInLoan(@Path("id") id: Long, @Body request: CheckInRequest = CheckInRequest()): Loan

    @DELETE("/api/loans/{id}")
    suspend fun deleteLoan(@Path("id") id: Long)

    @GET("/api/books/{id}/loans")
    suspend fun getBookLoans(@Path("id") bookId: Long): List<Loan>

    // Stats
    @GET("/api/stats")
    suspend fun getStats(): StatsResponse

    // Shelves
    @GET("/api/shelves")
    suspend fun listShelves(): List<Shelf>
}
