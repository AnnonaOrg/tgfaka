初始化运行环境
```bash 
sudo apt update && sudo apt -y upgrade && sudo apt install -y docker docker.io docker-compose &&  sudo systemctl enable docker && sudo systemctl start docker && sudo reboot
```
启动
```bash 
sudo docker-compose up -d
```
重启
```bash 
sudo docker-compose restart && sleep 10 && sudo docker-compose logs
```
重新编译
```bash 
sudo docker-compose build  && sudo docker-compose down  && sudo docker-compose up -d && sleep 10 && sudo docker-compose logs
```