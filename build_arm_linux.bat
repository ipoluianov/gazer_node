SET GOARCH=arm
SET GOOS=linux
go build -o gazer_node_arm_linux ./main/main.go
pause
