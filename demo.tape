Output demo.gif

Set Shell "bash"
Set FontFamily "JetBrains Mono"
Set FontSize 18
Set Width 1100
Set Height 720
Set Padding 24
Set TypingSpeed 50ms
Set Theme "Catppuccin Mocha"

Hide
Type "clear"
Enter
Show

Type "./bin/ssh-battos --local"
Enter
Sleep 2s

# Walk the page list down to GitHub
Type "j"
Sleep 500ms
Type "j"
Sleep 500ms
Type "j"
Sleep 500ms
Type "j"
Sleep 500ms
Type "j"
Sleep 700ms

# Open the GitHub repository list
Enter
Sleep 4s

# Filter down to this repository and open it
Type "/"
Sleep 600ms
Type "ssh.battos"
Sleep 800ms
Enter
Sleep 700ms
Enter
Sleep 4s

# README is selected by default. Focus the pane and scroll it.
Tab
Sleep 500ms
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms
Type "j"
Sleep 800ms

# Back to the file list, filter to a source file and open it
Tab
Sleep 500ms
Type "/"
Sleep 500ms
Type "main.go"
Sleep 700ms
Enter
Sleep 600ms
Enter
Sleep 2s

# Scroll the highlighted source file
Type "j"
Sleep 500ms
Type "j"
Sleep 800ms

# Back to the repository list, then to the portfolio pages, and quit
Type "Escape"
Sleep 1s
Type "Escape"
Sleep 1s
Type "q"
Sleep 1s
