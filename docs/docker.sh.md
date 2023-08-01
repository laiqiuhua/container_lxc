# About

docker.sh 是用 Shell 写的一个简易的 docker，支持以下功能：

* uts namespace
* mount namespace
* pid namespace
* network namespace
* memory 资源限制
* 联合加载
* 卷目录

《使用 Shell 脚本实现 Docker》讲解了以下功能：

* uts namespace
* mount namespace
* pid namespace
* network namespace
* memory 资源限制
* 联合加载
* 卷目录

《使用 Shell 脚本实现 Docker》讲解了以下功能：

* network namespace
* iptables

# Run

```
# 制作镜像
mkdir images
cd images
sudo debootstrap --arch amd64 xenial ./ubuntu1604

# 运行容器
sudo ./docker.sh -c run -m 100M -C dreamland -I ubuntu1604 -V data1 -P /bin/bash -n host -n none
```
