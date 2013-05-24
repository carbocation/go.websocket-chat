#scp forum.linux james@carbocation.com:/data/bin/upload.forum.linux
#ssh james@carbocation.com "sh -c 'cd /data/bin/; sudo service go-askbitcoin-upstart stop; mv upload.forum.linux forum.linux; sudo service go-askbitcoin-upstart start'"


scp chat.linux james@carbocation.com:/data/bin/upload.chat.linux
ssh -n -f james@carbocation.com "sh -c 'cd /data/bin/; killall chat.linux; mv upload.chat.linux chat.linux;GOMAXPROCS=8  nohup ./chat.linux > /dev/null 2>&1 &'"

