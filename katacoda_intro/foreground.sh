!#/bin/bash

n=1
GREEN='\033[0;32m'
NC='\033[0m' # No Color

while [ ! -f /root/post-run.log.done ] ;
do
      if test "$n" = "1"
      then
      clear
            n=$(( n+1 ))	 # increments $n
      else
      printf "."
      fi
      sleep 2
done
clear
echo -e "${GREEN}Ready to start your scenario${NC}"
