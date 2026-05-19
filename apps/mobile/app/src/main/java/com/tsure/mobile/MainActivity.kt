package com.tsure.mobile

import android.os.Bundle
import android.widget.Button
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity

class MainActivity : AppCompatActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        val subtitle = findViewById<TextView>(R.id.subtitleText)
        val actionButton = findViewById<Button>(R.id.primaryActionButton)

        subtitle.text = getString(R.string.dashboard_subtitle)
        actionButton.setOnClickListener {
            subtitle.text = getString(R.string.dashboard_ready_message)
        }
    }
}
