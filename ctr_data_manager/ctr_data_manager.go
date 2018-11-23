package ctr_data_manager

import "time"

import (
	"sync"
	"strings"
	"fandeco/ctr_server/db_conn"
	"fandeco/ctr_server/models"
	"fandeco/ctr_server/slogs"
	"fmt"
	"strconv"
	"reflect"
	"log"
)

var (
	CtrDM *ctrDataManager = nil
	mutex sync.Mutex
)

//метод вызывается при импорте пакета
func init() {
	if CtrDM == nil {
		CtrDM = new(ctrDataManager)
		CtrDM.initCtrLogs()
	}
}

func MapCopy(dst, src interface{}) {
	dv, sv := reflect.ValueOf(dst), reflect.ValueOf(src)

	for _, k := range sv.MapKeys() {
		dv.SetMapIndex(k, sv.MapIndex(k))
	}
}

type ctrDataManager struct {
	CtrDayLogs map[string]map[uint32]models.CtrCounter
}

func (ctrdm *ctrDataManager) GetСurDateString() (time.Time,string) {
	current_time := time.Now().Local()
	current_date_str := current_time.Format("2006-01-02")

	return current_time,current_date_str
}

func (ctrdm *ctrDataManager) GetYesterdayDateString() map[string]int {

	result_map := make(map[string]int)

	current_time := time.Now().Local()
	current_date_str := current_time.Format("2006-01-02")
	yesterday_date_str := current_time.Add(-24*time.Hour).Format("2006-01-02")
	result_map[current_date_str] = 1
	result_map[yesterday_date_str] = 1

	return result_map
}

func (ctdrm *ctrDataManager) RemoveOldData(){
	dates_map := ctdrm.GetYesterdayDateString()
	mutex.Lock()
	for key := range ctdrm.CtrDayLogs {
		_,ok := dates_map[key]
		if !ok{
			delete(ctdrm.CtrDayLogs, key)
		}
	}
	mutex.Unlock()
}



func (ctrdm *ctrDataManager) initCtrLogs()  {
	ctrdm.CtrDayLogs = make(map[string]map[uint32]models.CtrCounter)

	current_time,current_date_str := ctrdm.GetСurDateString()

	ctrdm.CtrDayLogs[current_date_str] = make(map[uint32]models.CtrCounter)

	rows, err := db_conn.Conn.Db.Query(`SELECT * FROM catalog_productctrcounter WHERE date=$1`,current_time)
	if err != nil {
		slogs.Logmg.Error.Println("Не смогли получить line - ", err.Error())
	} else {
		for rows.Next() {
			ctr_day_log := new(models.CtrCounter)

			err = rows.Scan(&ctr_day_log.Id,&ctr_day_log.Date,
				&ctr_day_log.Show,&ctr_day_log.Click,&ctr_day_log.Ctr,
				&ctr_day_log.ProductId,&ctr_day_log.CatalogClick)
			if err != nil {
				fmt.Printf("CtrCounter rows.Scan error: %v\n", err)
			}else{
				ctrdm.CtrDayLogs[current_date_str][ctr_day_log.ProductId] = *ctr_day_log
			}
		}
	}
}

func (ctrdm *ctrDataManager) IncrementCounter(ids string,counter_type string){
	if ids != "" {
		ids_list := strings.Split(ids,",")
		ids_list_number := []int{}
		for i := range ids_list {
			id_string := ids_list[i]
			if number ,err := strconv.Atoi(id_string); err == nil{
				ids_list_number = append(ids_list_number, number)
			}
		}
		//increment all clicks
		if len(ids_list_number) > 0 {
			current_time,cur_date_key := ctrdm.GetСurDateString()

			mutex.Lock()
			if _,ok := ctrdm.CtrDayLogs[cur_date_key]; ok{
			}else {
				ctrdm.CtrDayLogs[cur_date_key] = make(map[uint32]models.CtrCounter)
			}

			for i := range ids_list_number {
				product_id := uint32(ids_list_number[i])
				if product_data, id_ok := ctrdm.CtrDayLogs[cur_date_key][product_id];id_ok{
					if counter_type == "inc_click"{
						product_data.Click += 1
						product_data.Change = true
					} else if counter_type == "inc_show"{
						product_data.Show += 1
						product_data.Change = true
					}else if counter_type == "inc_catclick"{
						product_data.CatalogClick += 1
						product_data.Change = true
					}
					ctrdm.CtrDayLogs[cur_date_key][product_id] = product_data

				}else{
					ctr_day_log := models.CtrCounter{
						Id:0,
						ProductId:product_id,
						Date:current_time,
						Show:0,
						Click:0,
						CatalogClick:0,
						Ctr:0,
						Change:false,
					}
					if counter_type == "inc_click"{
						ctr_day_log.Click = 1
						ctr_day_log.Change = true
					} else if counter_type == "inc_show"{
						ctr_day_log.Show = 1
						ctr_day_log.Change = true
					} else if counter_type == "inc_catclick"{
						ctr_day_log.CatalogClick = 1
						ctr_day_log.Change = true
					}
					ctrdm.CtrDayLogs[cur_date_key][product_id] = ctr_day_log
				}
			}
			mutex.Unlock()
		}
	}
}

func (ctrdm *ctrDataManager) UpdateDataBase(date_str string){
	var current_date_str string

	if date_str == ""{
		_,current_date_str = ctrdm.GetСurDateString()
	}else{
		current_date_str = date_str
	}

	//update not existed counters
	ctrdm.CreateNotExistCounters(current_date_str)

	mutex.Lock()
	cur_day_data,ok :=  ctrdm.CtrDayLogs[current_date_str]
	mutex.Unlock()

	if ok {
		for _, v := range cur_day_data {
			_, err := db_conn.Conn.Db.Exec("UPDATE catalog_productctrcounter SET click = $1,catalog_click = $2, show = $3 WHERE product_id = $4 AND date = $5 ",v.Click,v.CatalogClick,v.Show,v.ProductId,current_date_str)
			if err != nil{
				slogs.Logmg.Error.Println("Не смогли обновить catalog_productctrcounter - ", err.Error())
			}
		}
	}
}

func (ctrdm *ctrDataManager) CreateNotExistCounters(date string){
	if cur_day_data,ok :=  ctrdm.CtrDayLogs[date];ok{

		stmt, err := db_conn.Conn.Db.Prepare("INSERT INTO catalog_productctrcounter (date, show, click, ctr, product_id,catalog_click) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id")
		if err != nil {
			log.Fatal(err)
		}

		defer stmt.Close()

		mutex.Lock()
		for _, v := range cur_day_data {
			if v.Id == 0{
				var counterId uint32
				err = stmt.QueryRow(date,v.Show,v.Click,v.Ctr,v.ProductId,v.CatalogClick).Scan(&counterId)
				if err != nil{
					slogs.Logmg.Error.Println("Не смогли создать новый объект catalog_productctrcounter - ", err.Error())
				}else{
					v.Id = counterId
					ctrdm.CtrDayLogs[date][v.ProductId] = v
				}
			}
		}
		mutex.Unlock()
	}
}