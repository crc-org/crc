# goautoit
golang invoke autoit function through AutoItX3.dll

## Install

```golang
go get -u github.com/shadow1163/goautoit
```

## Example

- open notepad
- type some string into notepad, eg: **"hello world"**
- close notepad without saving

```golang
goautoit.Run("notepad.exe")
goautoit.WinWait("Untitled")
goautoit.Send("hello world")
goautoit.WinClose("Untitled")
```
