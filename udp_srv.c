#include <stdio.h>  
#include <sys/types.h>  
#include <sys/socket.h>  
#include <netinet/in.h>  
#include <arpa/inet.h>  
#include <pthread.h>
#include <unistd.h>
#include <string.h>
#include <stdlib.h>
#include <sys/socket.h>

typedef struct {
    int arg1;
    int arg2;
} Args;

int Bind(int port) {
    printf("start bind\n");
    int server_sockfd;  
    struct sockaddr_in my_addr;   //服务器网络地址结构体  
    memset(&my_addr,0,sizeof(my_addr)); //数据初始化--清零  
    my_addr.sin_family=AF_INET; //设置为IP通信  
    my_addr.sin_addr.s_addr=INADDR_ANY;//服务器IP地址--允许连接到所有本地地址上  
    my_addr.sin_port=htons(port); //服务器端口号  

    /*创建服务器端套接字--IPv4协议，面向无连接通信，UDP协议*/  
    if((server_sockfd=socket(PF_INET,SOCK_DGRAM,0))<0)  
    {    
        perror("socket");  
        return -1;  
    }  

    int v = 1;
    setsockopt(server_sockfd, SOL_SOCKET, SO_REUSEPORT, &v, sizeof(v));

    /*将套接字绑定到服务器的网络地址上*/  
    if (bind(server_sockfd,(struct sockaddr *)&my_addr,sizeof(struct sockaddr))<0)  
    {  
        perror("bind");  
        return -1;  
    }  
    return server_sockfd;
}

void *thread(Args *arg) {
    int id = arg->arg1;
    int fd = Bind(3333);
    if (fd < 0) {
        printf("bind failed\n");
        exit(-1);
    }

    char buf[BUFSIZ];
    socklen_t sin_size=sizeof(struct sockaddr_in);
    struct sockaddr_in remote_addr;
    int len = 0;
    printf("start waiting for a packet...\n");  

    while (1) {
        if ((len = recvfrom(fd , buf, BUFSIZ, 0, 
                        (struct sockaddr *)&remote_addr, &sin_size)) < 0) {
            perror("recvfrom");
            return NULL; 
        } 
        printf("id:%d, received packet from %s:\n", id, inet_ntoa(remote_addr.sin_addr));  

    }
    close(fd);
}


int main(void) {  
    int i = 0;
    Args *arg = NULL;
    pthread_t pid;
    for (i = 0; i < 10; i++) {
        arg = (Args *)malloc(sizeof(Args));
        arg->arg1 = i;
        pthread_create(&pid, NULL, &thread, arg);
    }

    while (1) {
        sleep(100);
    }

}  

