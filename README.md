# CFR: Codeforces CLI Contest Helper

CFR is a fast, user-friendly CLI tool for competitive programmers who use Codeforces. It helps you download problems, organize your workspace, and test your solutions with ease—all from the command line!

## Features
- **Contest Loader:** Download all problems and sample tests for a contest in one command.
- **Organized Workspace:** Each problem gets its own folder with generic file names (`main.cpp`, `in.txt`, `out.txt`).
- **Language Support:** Works with C++, C, Rust, Go, Python, and Java (configurable).
- **Sample & Custom Testing:** Run all sample tests or your own custom test cases.
- **Persistent State:** Keeps track of loaded contests and problems in `.cfr/problems.json`.
- **Safe & Idempotent:** Prevents duplicate `init` or `load` commands.

---

## Quick Start

### 1. Install
Build the binary:
```sh
go build -o bin/cfr.exe
```

### 2. Initialize Your Workspace
Run this in your contest folder:
```sh
cfr init
```

### 3. Set Your Language
Edit `.cfr/config.json` and set your language, e.g.:
```json
{
  "language": "cpp"
}
```
Supported: `cpp`, `c`, `rust`, `go`, `python`, `java`

### 4. Load a Contest
```sh
cfr load <CONTEST_ID>
```
This will create a folder for each problem, e.g. `A. Sum of Round Numbers/`, with:
- `main.cpp` (or `main.<ext>`)
- `in.txt` (for custom input)
- `out.txt` (for custom output)
- All sample tests stored in `.cfr/problems.json`

### 5. Solve & Test
Write your solution in `main.cpp` (or the appropriate file).

#### Run All Sample Tests
```sh
cfr test <PROBLEM_ID>
```
Example:
```sh
cfr test A
```

#### Run a Custom Test
Edit `in.txt` in the problem folder, then:
```sh
cfr test -c <PROBLEM_ID>
```
The output will be written to `out.txt`.

---

## File Structure Example
```
YourContestFolder/
├── .cfr/
│   ├── config.json
│   └── problems.json
├── A. Sum of Round Numbers/
│   ├── main.cpp
│   ├── in.txt
│   └── out.txt
├── B. .../
│   ├── main.cpp
│   ├── in.txt
│   └── out.txt
...
```

---

## Tips
- You can re-run `cfr load <ID>` to update problems if needed.
- Only one contest can be loaded at a time per workspace.
- All state is stored in `.cfr/problems.json`.
- For C++/C/Rust/Go, the binary is built and run in the problem directory.

---

## Troubleshooting
- If you see errors about missing files or folders, make sure you ran `cfr init` and `cfr load <ID>`.
- If compilation fails, check your language setting in `.cfr/config.json` and your compiler installation.

---

## License
MIT

---

## Contributing
Pull requests and issues are welcome!
