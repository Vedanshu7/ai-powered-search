# AI-Powered Smart Search

A web application that uses OpenAI to understand natural language search queries and generates optimized Google search URLs. The application consists of a Go backend that processes queries through OpenAI's API and a React frontend for user interaction.

## Features

- Natural language query processing using OpenAI
- Advanced Google search parameter optimization
- File type filtering (PDF, DOC, etc.)
- Site-specific search options
- Date range filtering
- Real-time query processing
- Modern, responsive UI

## Tech Stack

- **Frontend:**
  - React
  - TailwindCSS
  - Lucide Icons

- **Backend:**
  - Go
  - OpenAI API

## Prerequisites

Before running this project, make sure you have:

- Node.js (v14 or higher)
- Go (v1.16 or higher)
- OpenAI API key

## Installation

### Backend Setup

1. Clone the repository:
```bash
git clone <repository-url>
cd your-project-name
```

2. Set up your OpenAI API key:
```code
const OPENAI_API_KEY="your-api-key"
```

3. Run the Go server:
```bash
go run main.go
```

The server will start on `http://localhost:8080`

### Frontend Setup

1. Navigate to the frontend directory:
```bash
cd smart-search
```

2. Install dependencies:
```bash
npm install
```

3. Start the development server:
```bash
npm start
```

The frontend will be available at `http://localhost:3000`

## Usage

1. Enter your search query in natural language (e.g., "find PDF research papers about machine learning from arxiv published in the last year")
2. Click the Search button
3. The application will:
   - Process your query using OpenAI
   - Generate an optimized search URL
   - Open the results in a new tab

## Project Structure

```
project-root/
├── backend/
│   └── main.go
└── smart-search/
    ├── public/
    ├── src/
    │   ├── components/
    │   │   └── SearchFrontend.js
    │   ├── App.js
    │   ├── index.js
    │   └── index.css
    ├── package.json
    └── tailwind.config.js
```

## Environment Variables

Backend:
- `OPENAI_API_KEY`: Your OpenAI API key
- `PORT`: Server port (default: 8080)

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details

## Acknowledgments

- OpenAI for providing the API
- React and Go communities for excellent documentation and tools