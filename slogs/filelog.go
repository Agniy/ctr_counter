/*
   Пакет создает систему логирования,доступ к логировнию мы получаем через вызов
   переменной
*/
package slogs

import (
	"log"
	"os"
	"fandeco/ctr_server/settings"
)

//структура singl,для создания объекта логера в файл
//---------------------------------
type single struct {
	Trace       *log.Logger //информация
	Info        *log.Logger //информация
	Warning     *log.Logger //предупреждение
	Error       *log.Logger //ошибка
	log_err_f   *os.File
	log_warn_f  *os.File
	log_info_f  *os.File
	log_trace_f *os.File
	Log_err     error
}

//метод инициализации настрок логирования
func (sl *single) initLogs() {

	//открываем все файлы
	//----------------------------------------------------------------------------------------
	sl.log_err_f, sl.Log_err = os.OpenFile(settings.LOG_ERR_FILE_PUTH, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if sl.Log_err != nil {
		panic(sl.Log_err.Error())
	}
	sl.log_warn_f, sl.Log_err = os.OpenFile(settings.LOG_WARN_FILE_PUTH, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if sl.Log_err != nil {
		panic(sl.Log_err.Error())
	}
	sl.log_info_f, sl.Log_err = os.OpenFile(settings.LOG_INFO_FILE_PUTH, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if sl.Log_err != nil {
		panic(sl.Log_err.Error())
	}
	sl.log_trace_f, sl.Log_err = os.OpenFile(settings.LOG_TRACE_FILE_PUTH, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if sl.Log_err != nil {
		panic(sl.Log_err.Error())
	}
	//----------------------------------------------------------------------------------------

	sl.Trace = log.New(sl.log_trace_f,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	sl.Info = log.New(sl.log_info_f,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	sl.Warning = log.New(sl.log_warn_f,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	sl.Error = log.New(sl.log_err_f,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

func (sl *single) CloseLogFiles() {
	sl.log_err_f.Close()
	sl.log_warn_f.Close()
	sl.log_info_f.Close()
	sl.log_trace_f.Close()
}

//---------------------------------

var (
	instance *single = nil
	Logmg    *single
)

//метод вызывается при импорте пакета
func init() {
	//создаем наш объект логирования
	Logmg = New()
}

//метод создания объекта логера
func New() *single {
	if instance == nil {
		instance = new(single)
		instance.initLogs()
	}
	return instance
}
