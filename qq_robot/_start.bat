@ECHO OFF

:: 修改console默认encoding为utf8，避免中文乱码
CHCP 65001

echo 启动机器人
start "qq_robot" C:\Windows\System32\cmd.exe /k "D:\_codes\Python\dnf_calc\robot\qq_robot.exe"
