# SplitBill AI

SplitBill AI is a modern web application designed to simplify splitting bills among friends. It leverages an LLM-powered backend workflow to automatically extract items and prices from a receipt image, allowing users to quickly assign items and calculate who owes what.

The project is built with a decoupled frontend and backend architecture, making it scalable and easy to maintain.

### ✨ Key Features

-   **Image Upload**: Snap a picture of your receipt to start.
-   **AI-Powered OCR**: Automatically scans the receipt and structures it into a list of items and prices using an n8n + LLM workflow.
-   **Interactive Item Assignment**: Easily assign each item to one or more people in your group.
-   **Real-Time Calculation**: Instantly see the total amount each person owes.
-   **Session-Based**: No login required. Just upload, split, and share.

### 🛠️ Tech Stack

-   **Backend**: **Go (Golang)** with the **Gin** framework.
-   **Frontend**: **Next.js** (React).
-   **Workflow Automation**: **n8n.io**.
-   **AI / OCR**: **Large Language Model** (Gemini or GPT) for receipt parsing.
-   **Database**: **PostgreSQL**.

### 🚀 Getting Started

1.  Clone the repository: `git clone -`
2.  Navigate to the `splitbill-llmocr-api` directory and run `go run main.go`.
3.  Navigate to the `splitbill-llmocr-web` directory and run `npm run dev`.
4.  Set up your n8n workflow and configure the webhook URL in the backend.