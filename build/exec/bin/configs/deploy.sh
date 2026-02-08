#!/bin/bash

SERVERS=("svr" "backendsvr" "apisvr" "usersvr" "centersvr" "sportsvr")
VALID_SERVERS=(${SERVERS[*]:1:${#SERVERS[@]}-1})
echo -e "#  \033[35m [${VALID_SERVERS[*]}] \033[0m"
ACTIONS=("inst" "uninst" "start" "stop" "restart" "ps" "reload") # "up"
MIN_SERVER_ID=10
MAX_SERVER_ID=12

# shellcheck disable=SC2006
cur_path=$(pwd)
basename=$0
basedir=$(
  cd "$(dirname $basename)"
  pwd
)

cd $basedir

echo -e "#  \033[35m ========================================================================================= \033[0m"
echo -e "#  \033[32m WELCOME TO USE [\033[31m$basename $*\033[0m\033[32m]!\033[0m"
echo -e "#  \033[34m basedir => \033[33m$basedir\033[0m\033[34m curdir => \033[33m$cur_path\033[0m"
echo "# "
echo -e "#  \033[35m ------------------------------------ \033[0m"
# shellcheck disable=SC2021
action=$(echo $1 | tr "[A-Z]" "[a-z]")
# shellcheck disable=SC2021
wuhan=$(echo $2 | tr "[A-Z]" "[a-z]")

# shellcheck disable=SC2070
if [ -n "$server" ] && [[ ! $server =~ svr$ ]]; then
  echo -e "#  \033[32m auto add 'svr' for second argument 'server' \033[0m"
  # shellcheck disable=SC2116
  wuhan=$(echo "${wuhan}svr")
fi

sid=$3
echo -e "#  \033[34m Action => \033[33m$action\033[0m"
echo -e "#  \033[34m Server => \033[33m$server $sid \033[0m"
echo -e "#  \033[35m ------------------------------------ \033[0m"

exitFn() {
  echo "# "
  if [ $# -gt 0 ]; then
    echo -e "#  \033[31m deploy error: $* \033[0m"
    echo -e "#  \033[31m deploy failed, check for error! \033[0m"
  else
    echo -e "#  \033[32m deploy succeed! \033[0m"
  fi
  cd $cur_path
  echo "# "
  echo -e "#  \033[35m ========================================================================================= \033[0m"
  exit
}

untilStoppedFn() {
  local arg_server_name=$1
  local arg_server_id=$2
  # shellcheck disable=SC2116
  local arg_server=$(echo "$arg_server_name $arg_server_id")
  echo "# "
  echo -e "#  \033[34m ------------------------------------waiting for $arg_server stop start------------------------------------ \033[0m"

  local tryTimes=0

  while :; do
    tryTimes=$(expr $tryTimes + 1)
    if [ $tryTimes -gt 30 ]; then
      exitFn "waiting for $arg_server stop timeout...150s, retry later!"
    fi
    local proc_num=0
    if [ $arg_server_name = "sportsvr" ] && [ ${#arg_server_id} -gt 0 ]; then
      proc_num=$(ps -ef | grep "$arg_server" | grep -v grep | grep -v $basename | wc -l)
    else
      proc_num=$(ps -ef | grep $arg_server | grep -v grep | grep -v $basename | wc -l)
    fi
    echo "#   ($tryTimes/30): waiting stop proc_num => $proc_num ..."
    if [ $proc_num -gt 0 ]; then
      echo -e "#  \033[36m"
      if [ $arg_server_name = "sportsvr" ] && [ ${#arg_server_id} -gt 0 ]; then
        ps -ef | grep "$arg_server" | grep -v grep | grep -v $basename
      else
        ps -ef | grep $arg_server | grep -v grep | grep -v $basename
      fi
      echo -e "\033[0m#   please wait..."
      echo "#  "
      sleep 5
    else
      echo "#   waiting for $arg_server stop done"
      break
    fi
  done

  echo "# "
  echo -e "#  \033[34m ------------------------------------waiting for $arg_server stop completed------------------------------------ \033[0m"
}

checkArgsFn() {
  # -z表示后面的值是否为空，为空则返回true，否则返回false。
  # -n表示判断后面的值是否为空，不为空则返回true，为空则返回false。
  echo "# "
  echo -e "#  \033[34m check argument... \033[0m"

  install_flag=false

  if [ -z "$action" ]; then
    exitFn "The first argument 'action' must be not null!"
  else
    flag=false
    for i in "${ACTIONS[@]}"; do
      if [ "$i" = "$action" ]; then
        if [ "$i" = "inst" ]; then
          install_flag=true
        fi
        flag=true
        break
      fi
    done

    if [ "$flag" = false ]; then
      # shellcheck disable=SC2145
      exitFn "The first argument 'action' must in \033[33m'${ACTIONS[@]}'\033[0m any one, but got \033[33m'$action'\033[0m"
    fi
  fi

  if [ -z "$server" ]; then
    wuhan="svr"
  else
    if [ "$install_flag" = true ]; then
      echo -e "#  \033[33m When the first argument 'action'=inst uninst up, the other arguments are unnecessary\033[0m"
    else
      flag=false
      game_flag=false
      for i in "${SERVERS[@]}"; do
        if [ "$i" = "$server" ]; then
          if [ "$i" = "sportsvr" ]; then
            game_flag=true
          fi
          flag=true
          break
        fi
      done

      if [ "$flag" = false ]; then
        # shellcheck disable=SC2145
        exitFn "The second argument 'server' must in \033[33m'${SERVERS[@]}'\033[0m any one or\033[33m null\033[0m, but got \033[33m'$server'\033[0m"
      fi
    fi
  fi
  if [ "$game_flag" = true ] || [ "$action" = "up" ] || [ "$action" = "start" ] || [ "$action" = "restart" ]; then
    if [ -z "$sid" ]; then
      if [ -z "$MAX_GAME_SERVER_ID" ]; then
        echo -e "#  \033[1;41;33mThe third argument 'gamesvr_id' is null and the system env 'MAX_GAME_SERVER_ID' is NULL! \033[0m"
      else
        if grep '^[[:digit:]]*$' <<<"$MAX_GAME_SERVER_ID"; then
          echo -e "#  \033[34m check 'MAX_GAME_SERVER_ID':\033[31m$sid\033[0m \033[34mOK!\033[0m"
        else
          exitFn "The system env 'MAX_GAME_SERVER_ID' must be a number!"
        fi
        MAX_SERVER_ID=$MAX_GAME_SERVER_ID
        if [ $MAX_SERVER_ID -lt $MIN_SERVER_ID ]; then
          echo -e "#  \033[34m MAX_GAME_SERVER_ID($MAX_SERVER_ID) < $MIN_SERVER_ID => MAX_GAME_SERVER_ID = $MIN_SERVER_ID\033[0m"
          MAX_SERVER_ID=$MIN_SERVER_ID
        fi
      fi
      echo "# "
      echo -e "#  \033[1;41;33m Range of game server ids is $MIN_SERVER_ID - $MAX_SERVER_ID \033[0m"
      echo "# "
    else
      if grep '^[[:digit:]]*$' <<<"$sid"; then
        echo -e "#  \033[34m check 'gamesvr_id':\033[31m$sid\033[0m \033[34mOK!\033[0m"
      else
        exitFn "The third argument 'gamesvr_id' must be a number!"
      fi
    fi
  fi
  if [ $server = "svr" ]; then
    sid=""
  fi
}

installFn() {
  echo "# "
  echo -e "#  \033[34m ------------------------------------install start------------------------------------ \033[0m"
  echo -e "#  \033[35m (1/6): check installed \033[0m"
  if [ -e "bin" ]; then
    exitFn "already installed, you can't install now"
  fi
  echo -e "#  \033[35m (2/6): check server running \033[0m"
  local proc_num=$(ps -ef | grep svr | grep -v grep | grep -v $basename | wc -l)
  echo "#   proc_num => $proc_num"
  # shellcheck disable=SC2086
  if [ $proc_num -gt 0 ]; then
    exitFn "server is running, you can't install now"
  fi

  mkdir bin

  echo -e "#  \033[35m (3/6): check zip file exist \033[0m"
  for f in ./bin_*.7z; do
    if [ -e "$f" ]; then
      echo -e "#  \033[34m find 7z file in './bin/' named: \033[32m$f \033[0m"
      echo -e "#  \033[35m (4/6): remove old files and programs \033[0m"
      #      cd bin && ls | egrep -v .7z | xargs rm -rf && cd .. || exitFn "remove old files and programs error"
      echo -e "#  \033[35m (5/6): install 7z commend \033[0m"
      yum install p7zip p7zip-plugins || exitFn "install 7z error"
      echo -e "#  \033[35m (6/6): unzip 7z file \033[0m"
      7z x "$f" -r -aoa -o./bin || exitFn "unzip 7z file error"
    else
      exitFn "7z file not exist in './bin/', please check it!"
    fi
    break
  done
  echo "# "
  echo -e "#  \033[34m ------------------------------------install completed------------------------------------ \033[0m"
  sleep 1
}

startFn() {
  local arg_server_name=$1
  local arg_server_id=$2
  # shellcheck disable=SC2116
  local arg_server=$(echo "$arg_server_name $arg_server_id")
  # shellcheck disable=SC2116
  local exec_path=$(echo "bin/cmd/$arg_server_name/")
  echo "# "
  echo -e "#  \033[34m ------------------------------------start $arg_server start------------------------------------ \033[0m"
  echo "#   pwd=$(pwd)"
  if test -e "$exec_path$arg_server_name"; then
    echo "#   found exec file => $exec_path$arg_server_name"
    local proc_num=0
    if [ $arg_server_name = "sportsvr" ] && [ ${#arg_server_id} -gt 0 ]; then
      proc_num=$(ps -ef | grep "$arg_server" | grep -v grep | grep -v $basename | wc -l)
    else
      proc_num=$(ps -ef | grep $arg_server | grep -v grep | grep -v $basename | wc -l)
    fi
    echo "#   proc_num => $proc_num"
    # shellcheck disable=SC2086
    if [ $proc_num -le 0 ]; then
      cd $exec_path
      chmod a+x $arg_server_name
      nohup ./$arg_server &
      cd $basedir
      echo -e "#  \033[33m $arg_server startup and running\033[0m"
    else
      exitFn "$arg_server is running, you have to stop it first!"
    fi
  else
    exitFn "$exec_path$arg_server_name not found, you have to use first argument 'action=inst' or 'action=up' to install it first."
  fi

  echo "# "
  echo -e "#  \033[34m ------------------------------------start $arg_server completed------------------------------------ \033[0m"
  sleep 1
}

stopFn() {
  local arg_server_name=$1
  local arg_server_id=$2
  # shellcheck disable=SC2116
  local arg_server=$(echo "$arg_server_name $arg_server_id")
  echo "# "
  echo -e "#  \033[34m ------------------------------------stop $arg_server start------------------------------------ \033[0m"
  # shellcheck disable=SC2006
  # shellcheck disable=SC2126
  # shellcheck disable=SC2009
  local proc_num=0
  if [ $arg_server_name = "sportsvr" ] && [ ${#arg_server_id} -gt 0 ]; then
    proc_num=$(ps -ef | grep "$arg_server" | grep -v grep | grep -v $basename | wc -l)
  else
    proc_num=$(ps -ef | grep $arg_server | grep -v grep | grep -v $basename | wc -l)
  fi
  echo "#   proc_num => $proc_num"
  # shellcheck disable=SC2086
  if [ $proc_num -le 0 ]; then
    exitFn "$arg_server is not running, you can't stop it!"
  else
    if [ $arg_server_name = "sportsvr" ] && [ ${#arg_server_id} -gt 0 ]; then
      ps -aux | grep "$arg_server" | grep -v grep | grep -v $basename | awk '{print $2}' | xargs kill -2 && echo -e "#  \033[33m $arg_server stop OK\033[0m" || exitFn "$arg_server stop failed!"
    else
      ps -aux | grep $arg_server | grep -v grep | grep -v $basename | awk '{print $2}' | xargs kill -2 && echo -e "#  \033[33m $arg_server stop OK\033[0m" || exitFn "$arg_server stop failed!"
    fi
  fi

  untilStoppedFn $arg_server_name $arg_server_id

  echo "# "
  echo -e "#  \033[34m ------------------------------------stop $arg_server completed------------------------------------ \033[0m"
  sleep 1
}

psFn() {
  local arg_server_name=$1
  local arg_server_id=$2
  # shellcheck disable=SC2116
  local arg_server=$(echo "$arg_server_name $arg_server_id")
  echo "# "
  echo -e "#  \033[34m ------------------------------------ps $arg_server start------------------------------------ \033[0m"
  echo "# "
  # shellcheck disable=SC2009
  ps aux | grep $arg_server | grep -v grep | grep -v $basename
  echo "# "
  echo -e "#  \033[34m ------------------------------------ps $arg_server completed------------------------------------ \033[0m"
}

restartFn() {
  local arg_server_name=$1
  local arg_server_id=$2
  # shellcheck disable=SC2116
  local arg_server=$(echo "$arg_server_name $arg_server_id")
  # shellcheck disable=SC2116
  local exec_path=$(echo "bin/cmd/$arg_server_name/")
  echo "# "
  echo -e "#  \033[34m ------------------------------------restart $arg_server start------------------------------------ \033[0m"
  echo "#   pwd=$(pwd)"
  if test -e "$exec_path$arg_server_name"; then
    echo "#   (1/3): found exec file => $exec_path$arg_server_name"
    echo "#   (2/3): stop process"
    # shellcheck disable=SC2006
    # shellcheck disable=SC2126
    # shellcheck disable=SC2009
    local proc_num=0
    if [ $arg_server_name = "sportsvr" ] && [ ${#arg_server_id} -gt 0 ]; then
      proc_num=$(ps -ef | grep "$arg_server" | grep -v grep | grep -v $basename | wc -l)
    else
      proc_num=$(ps -ef | grep $arg_server | grep -v grep | grep -v $basename | wc -l)
    fi
    echo "#   proc_num => $proc_num"
    # shellcheck disable=SC2086
    if [ $proc_num -gt 0 ]; then
      if [ $arg_server_name = "sportsvr" ] && [ ${#arg_server_id} -gt 0 ]; then
        ps -aux | grep "$arg_server" | grep -v grep | grep -v $basename | awk '{print $2}' | xargs kill -2 && echo -e "#  \033[33m $arg_server stop OK\033[0m" || exitFn "$arg_server stop failed!"
      else
        ps -aux | grep $arg_server | grep -v grep | grep -v $basename | awk '{print $2}' | xargs kill -2 && echo -e "#  \033[33m $arg_server stop OK\033[0m" || exitFn "$arg_server stop failed!"
      fi
    fi
    untilStoppedFn $arg_server_name $arg_server_id
    echo "#   (3/3): startup process"
    cd $exec_path
    chmod a+x "$arg_server_name"
    nohup ./"$arg_server_name" "$arg_server_id" &
    cd $basedir
    echo -e "#  \033[33m $arg_server startup and running\033[0m"
  else
    exitFn "$exec_path not found, you have to use first argument 'action=inst' or 'action=up' to install it first."
  fi

  echo "# "
  echo -e "#  \033[34m ------------------------------------restart $arg_server completed------------------------------------ \033[0m"
  sleep 1
}

reloadFn() {
  local arg_server_name=$1
  local arg_server_id=$2
  # shellcheck disable=SC2116
  local arg_server=$(echo "$arg_server_name $arg_server_id")
  echo "# "
  echo -e "#  \033[34m ------------------------------------reload $arg_server start------------------------------------ \033[0m"
  # shellcheck disable=SC2006
  # shellcheck disable=SC2126
  # shellcheck disable=SC2009
  local proc_num=0
  if [ $arg_server_name = "sportsvr" ] && [ ${#arg_server_id} -gt 0 ]; then
    proc_num=$(ps -ef | grep "$arg_server" | grep -v grep | grep -v $basename | wc -l)
  else
    proc_num=$(ps -ef | grep $arg_server | grep -v grep | grep -v $basename | wc -l)
  fi
  echo "#   proc_num => $proc_num"
  # shellcheck disable=SC2086
  if [ $proc_num -le 0 ]; then
    exitFn "$arg_server is not running, you can't reload for it!"
  else
    if [ $arg_server_name = "sportsvr" ] && [ ${#arg_server_id} -gt 0 ]; then
      ps -aux | grep grep "$arg_server" | grep -v grep | grep -v $basename | awk '{print $2}' | xargs kill -5 && echo -e "#  \033[33m$arg_server reload OK\033[0m" || exitFn "$arg_server reload failed!"
    else
      ps -aux | grep grep $arg_server | grep -v grep | grep -v $basename | awk '{print $2}' | xargs kill -5 && echo -e "#  \033[33m$arg_server reload OK\033[0m" || exitFn "$arg_server reload failed!"
    fi
  fi
  echo "# "
  echo -e "#  \033[34m ------------------------------------reload $arg_server completed------------------------------------ \033[0m"
  sleep 1
}

killFn() {
  local arg_server_name=$1
  local arg_server_id=$2
  # shellcheck disable=SC2116
  local arg_server=$(echo "$arg_server_name $arg_server_id")
  echo "# "
  echo -e "#  \033[34m ------------------------------------kill $arg_server start------------------------------------ \033[0m"
  # shellcheck disable=SC2006
  # shellcheck disable=SC2126
  # shellcheck disable=SC2009
  local proc_num=0
  if [ $arg_server_name = "sportsvr" ] && [ ${#arg_server_id} -gt 0 ]; then
    proc_num=$(ps -ef | grep "$arg_server" | grep -v grep | grep -v $basename | wc -l)
  else
    proc_num=$(ps -ef | grep $arg_server | grep -v grep | grep -v $basename | wc -l)
  fi
  echo "#   proc_num => $proc_num"
  # shellcheck disable=SC2086
  if [ $proc_num -le 0 ]; then
    exitFn "$arg_server is not running, you can't kill it!"
  else
    if [ $arg_server_name = "sportsvr" ] && [ ${#arg_server_id} -gt 0 ]; then
      ps -aux | grep "$arg_server" | grep -v grep | grep -v $basename | awk '{print $2}' | xargs kill -9 && echo -e "#  \033[33m$arg_server kill OK\033[0m" || exitFn "$arg_server kill failed!"
    else
      ps -aux | grep $arg_server | grep -v grep | grep -v $basename | awk '{print $2}' | xargs kill -9 && echo -e "#  \033[33m$arg_server kill OK\033[0m" || exitFn "$arg_server kill failed!"
    fi
  fi
  untilStoppedFn $arg_server_name $arg_server_id
  echo "# "
  echo -e "#  \033[34m ------------------------------------kill $arg_server completed------------------------------------ \033[0m"
  sleep 1
}

uninstallFn() {
  echo "# "
  echo -e "#  \033[34m ------------------------------------uninstall start------------------------------------ \033[0m"
  echo -e "#  \033[35m (1/2): stop all programs \033[0m"
  # shellcheck disable=SC2006
  # shellcheck disable=SC2126
  # shellcheck disable=SC2009
  local proc_num=$(ps -ef | grep svr | grep -v grep | grep -v $basename | wc -l)
  echo "#   proc_num => $proc_num"
  # shellcheck disable=SC2086
  if [ $proc_num -gt 0 ]; then
    ps -aux | grep svr | grep -v grep | grep -v $basename | awk '{print $2}' | xargs kill -2 && echo -e "#  \033[33m all servers stop OK\033[0m" || exitFn "all servers stop failed!"
  fi
  untilStoppedFn "svr"
  echo -e "#  \033[35m (2/2): remove old files and programs \033[0m"
  #  cd bin && ls | egrep -v .7z | xargs rm -rf && cd .. || exitFn "remove old files and programs error"
  rm -rf bin
  echo "# "
  echo -e "#  \033[34m ------------------------------------uninstall completed------------------------------------ \033[0m"
}

#updateFn(){
#  echo "# "
#  echo -e "#  \033[34m ------------------------------------update start------------------------------------ \033[0m"
#  # shellcheck disable=SC2006
#  # shellcheck disable=SC2126
#  # shellcheck disable=SC2009
#  local proc_num=`ps -ef | grep svr | grep -v grep|grep -v $basename | wc -l`
#  echo "#   proc_num => $proc_num"
#  # shellcheck disable=SC2086
#  if [ $proc_num -gt 0 ] ;then
#    ps -aux |grep svr|grep -v grep|grep -v $basename|awk '{print $2}'|xargs kill -2 && echo -e "#  \033[33m all servers stop OK\033[0m" || exitFn "all servers stop failed!"
#  fi
#
#  untilStoppedFn "svr"
#
#  cd bin && ls | egrep -v .7z | xargs rm -rf && cd .. || exitFn "remove old files and programs error"
#
#  git pull && installFn || exitFn "git pull latest version failed! "
#
#  for s in "${VALID_SERVERS[@]}";
#  do
#    if [ "$s" = "sportsvr" ]
#    then
#      # shellcheck disable=SC2004
#      for ((i = ${MIN_SERVER_ID}; i <= ${MAX_SERVER_ID}; i++))
#      do
#        echo "#   start.. $s $i"
#        startFn "$s" $i
#      done
#    else
#      echo "#   start.. $s"
#      startFn "$s"
#    fi
#  done
#
#  echo "# "
#  echo -e "#  \033[34m ------------------------------------update completed------------------------------------ \033[0m"
#}

sleep 1

checkArgsFn

case $action in
"inst")
  installFn
  ;;
"uninst")
  uninstallFn
  ;;
  #"up")
  #  updateFn
  #  ;;
"ps")
  psFn $server $sid
  ;;
"stop")
  stopFn $server $sid
  ;;
"reload")
  reload $server $sid
  ;;
"kill")
  killFn $server $sid
  ;;
"start")
  if [ -z "$server" ] || [ $server = "svr" ] || [ $server = "" ] || [ $server = "*" ]; then
    for s in "${VALID_SERVERS[@]}"; do
      if [ "$s" = "sportsvr" ]; then
        # shellcheck disable=SC2004
        for ((i = ${MIN_SERVER_ID}; i <= ${MAX_SERVER_ID}; i++)); do
          echo "#   start.. $s $i"
          startFn "$s" $i
        done
      else
        echo "#   start.. $s"
        startFn "$s"
      fi
    done
  elif [ $server = "sportsvr" ]; then
    if [ -z $sid ] || [ $sid = "" ] || [ $sid = "0" ] || [ $sid = 0 ]; then
      for ((i = ${MIN_SERVER_ID}; i <= ${MAX_SERVER_ID}; i++)); do
        echo "#   start.. $server $i"
        startFn $server $i
      done
    else
      if [ $sid -lt $MIN_SERVER_ID ] || [ $sid -gt $MAX_SERVER_ID ]; then
        exitFn "the third argument 'gamesvr_id'($sid) is out of the range $MIN_SERVER_ID - $MAX_SERVER_ID"
      else
        echo "#   start.. $server $sid"
        startFn $server @sid
      fi
    fi
  else
    echo "#   start.. $server"
    startFn $server
  fi
  ;;
"restart")
  if [ -z "$server" ] || [ $server = "svr" ] || [ $server = "" ] || [ $server = "*" ]; then
    for s in "${VALID_SERVERS[@]}"; do
      if [ "$s" = "sportsvr" ]; then
        # shellcheck disable=SC2004
        for ((i = ${MIN_SERVER_ID}; i <= ${MAX_SERVER_ID}; i++)); do
          echo "#   restart.. $s $i"
          restartFn "$s" $i
        done
      else
        echo "#   restart.. $s"
        restartFn "$s"
      fi
    done
  elif [ $server = "sportsvr" ]; then
    if [ -z $sid ] || [ $sid = "" ] || [ $sid = "0" ] || [ $sid = 0 ]; then
      for ((i = ${MIN_SERVER_ID}; i <= ${MAX_SERVER_ID}; i++)); do
        echo "#   restart.. $server $i"
        restartFn $server $i
      done
    else
      if [ $sid -lt $MIN_SERVER_ID ] || [ $sid -gt $MAX_SERVER_ID ]; then
        exitFn "the third argument 'gamesvr_id'($sid) is out of the range $MIN_SERVER_ID - $MAX_SERVER_ID"
      else
        echo "#   restart.. $server $sid"
        restartFn $server $sid
      fi
    fi
  else
    echo "#   restart.. $server"
    restartFn $server
  fi
  ;;
esac

cd $cur_path
echo "# "
echo -e "#  \033[35m ========================================================================================= \033[0m"
