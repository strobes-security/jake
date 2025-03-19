# Jake - Handle Hijack Checker

A fast and concurrent CLI tool by **Strobes Security** to scan web pages and their common subpages for social media handles (Twitter, LinkedIn, YouTube, Facebook, Instagram, TikTok) and detect potentially hijackable Twitter handles.

---

## ✨ Features
- Scans main URLs and common sub-pages (like `/contact`, `/about-us`, `/team`, etc.)
- Detects social media handles via regex
- Checks Twitter handle availability to detect hijackable accounts
- Multi-threaded concurrent scanning
- Outputs structured JSON results
- Verbose mode for detailed logging

---

## 🚀 Installation

Make sure you have **Go** installed.  
Then, run:

```bash
go install github.com/strobes-security/jake@latest
```

This will download and install the binary to your `$GOPATH/bin` or `$HOME/go/bin`.

---

## 🔎 Usage

### Basic usage:
```bash
jake -f urls.txt
```

### Advanced usage with all flags:
```bash
jake -f urls.txt -t 10 -o result.json -v
```

---

## 📜 Command-line Flags

| Flag            | Shorthand | Description                                   | Default        |
|-----------------|-----------|-----------------------------------------------|----------------|
| `--file`        | `-f`      | Path to a file containing URLs (required)     |                |
| `--threads`     | `-t`      | Number of concurrent workers                  | 5              |
| `--output`      | `-o`      | Output JSON file path                         | `result.json`  |
| `--verbose`     | `-v`      | Enable verbose output                         | `false`        |

---

## 📂 Example Input File (urls.txt)
```
https://example.com
https://anotherexample.org
https://companysite.io
```

---

## ✅ Output Format (result.json)
```json
[
  {
    "url": "https://example.com",
    "handles": [
      {
        "platform": "twitter",
        "handle": "example_handle",
        "hijackable": false
      },
      {
        "platform": "linkedin",
        "handle": "example-profile",
        "hijackable": false
      }
    ]
  },
  {}
]
```

---

## 💡 Todo Ideas for Future Improvements
- Add proxy support
- Timeout handling and retries
- Check hijackability for other platforms
- Export to CSV format

---

## 📃 License
This project is licensed under the MIT License.

---

## 🤝 Contributing
Pull requests are welcome! Please open an issue first if you’d like to suggest major changes.

---

## 👤 Author
Made with ❤️ by **Strobes Security** (https://github.com/strobes-security)
