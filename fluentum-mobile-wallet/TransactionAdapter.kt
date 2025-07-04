package com.fluentum.wallet.ui

import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.TextView
import androidx.recyclerview.widget.RecyclerView
import com.fluentum.wallet.R
import com.fluentum.wallet.data.Transaction
import com.fluentum.wallet.data.TransactionStatus
import java.text.SimpleDateFormat
import java.util.*

class TransactionAdapter(
    private var transactions: List<Transaction>,
    private val onTransactionClick: (Transaction) -> Unit
) : RecyclerView.Adapter<TransactionAdapter.TransactionViewHolder>() {

    private val dateFormat = SimpleDateFormat("MMM dd, yyyy HH:mm", Locale.getDefault())

    class TransactionViewHolder(itemView: View) : RecyclerView.ViewHolder(itemView) {
        val tvAmount: TextView = itemView.findViewById(R.id.tvAmount)
        val tvAddress: TextView = itemView.findViewById(R.id.tvAddress)
        val tvDate: TextView = itemView.findViewById(R.id.tvDate)
        val tvStatus: TextView = itemView.findViewById(R.id.tvStatus)
        val tvHash: TextView = itemView.findViewById(R.id.tvHash)
    }

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): TransactionViewHolder {
        val view = LayoutInflater.from(parent.context)
            .inflate(R.layout.item_transaction, parent, false)
        return TransactionViewHolder(view)
    }

    override fun onBindViewHolder(holder: TransactionViewHolder, position: Int) {
        val transaction = transactions[position]
        
        holder.tvAmount.text = transaction.formattedAmount
        holder.tvAddress.text = formatAddress(transaction)
        holder.tvDate.text = dateFormat.format(transaction.date)
        holder.tvStatus.text = transaction.status.getDisplayName()
        holder.tvStatus.setTextColor(transaction.status.getColor())
        holder.tvHash.text = transaction.hash.take(8) + "..."
        
        holder.itemView.setOnClickListener {
            onTransactionClick(transaction)
        }
    }

    override fun getItemCount(): Int = transactions.size

    fun updateTransactions(newTransactions: List<Transaction>) {
        transactions = newTransactions
        notifyDataSetChanged()
    }

    private fun formatAddress(transaction: Transaction): String {
        return when {
            transaction.amount > 0 -> "From: ${transaction.from.take(10)}..."
            else -> "To: ${transaction.to.take(10)}..."
        }
    }
} 