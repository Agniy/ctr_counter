package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"runtime"

	"fandeco/ctr_server/db_conn"
	"fandeco/ctr_server/slogs"
	"fandeco/ctr_server/ctr_data_manager"
)

const (
	CONN_HOST = "localhost"
	CONN_PORT = "5959"
	CONN_TYPE = "tcp"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() - 1)

	//Устанавливаем подключение к порту
	//------------------------------------------------------
	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	//------------------------------------------------------

	//отрабатываем остановку сервера
	//-------------------------------
	defer db_conn.Conn.Db.Close()
	defer l.Close() //Закрываем подключение
	defer slogs.Logmg.CloseLogFiles()
	//-------------------------------

	fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)
	for {
		//Следим за входящими подключениями
		conn, err := l.Accept()
		if err != nil {
			slogs.Logmg.Error.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		//Обрабатываем входящее подключение
		go handleRequest(conn)
	}
}

//echo '{"action":"inc_click","ids":"16609,23554"}' | nc localhost 5555
//echo '{"action":"inc_show","ids":"16609,23554"}' | nc localhost 5555
//echo '{"action":"update_db"}' | nc localhost 5555

// Обработка входящих соединений
func handleRequest(conn net.Conn) {
	var (
		json_data map[string]interface{} //словарь с данными
	)

	//декодируем данные в json формат
	//------------------------------------------------
	d := json.NewDecoder(conn)
	err := d.Decode(&json_data)
	if err != nil {
		slogs.Logmg.Error.Println("Не смогли разобрать присланый json - ", err.Error())
		conn.Write([]byte("Error:wrong json string format"))
	} else {
		if action,ok := json_data["action"]; ok {
			if action == "inc_click" || action == "inc_show" || action == "inc_catclick"{
				if ids,ok_ids := json_data["ids"];ok_ids{
					ids_string,ids_convert_ok := ids.(string)
					action_string,action_convert_ok := action.(string)

					if ids_convert_ok && action_convert_ok {
						ctr_data_manager.CtrDM.IncrementCounter(ids_string, action_string)
						conn.Write([]byte("ok"))
					}else{
						slogs.Logmg.Error.Println("ids или action в json_data на строка!")
					}
				}
			}else if action == "update_db"{
				//Try to get date str from parametr
				if date_str,date_str_ok := json_data["date_str"]; date_str_ok{
					if date_str_string,date_str_convert_ok := date_str.(string);date_str_convert_ok{
						ctr_data_manager.CtrDM.UpdateDataBase(date_str_string)
					}
				}else{
					ctr_data_manager.CtrDM.UpdateDataBase("")
				}
				conn.Write([]byte("ok"))
			}else if action == "update_not_exist" {
				_, current_date_str := ctr_data_manager.CtrDM.GetСurDateString()
				ctr_data_manager.CtrDM.CreateNotExistCounters(current_date_str)
				conn.Write([]byte("ok"))
			}else if action == "remove_old_data"{
				ctr_data_manager.CtrDM.RemoveOldData()
			}else{
				conn.Write([]byte("Error:not valid action parameter"))
			}
		} else{
			conn.Write([]byte("Error:no action parameter"))
		}
	}
	// Закрываем соединение после того, как все сделаем
	conn.Close()
}
