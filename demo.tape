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

# Walk the page list down to Projects
Type "j"
Sleep 700ms
Type "j"
Sleep 700ms
Type "j"
Sleep 700ms

# Open the selected page
Enter
Sleep 1s

# Scroll through the content
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms
Type "j"
Sleep 1s

# Back to the list, then quit
Tab
Sleep 700ms
Type "q"
Sleep 1s
