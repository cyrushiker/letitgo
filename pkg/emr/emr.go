package emr

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"

	_ "github.com/go-sql-driver/mysql" // mysql driver
)

type tables struct {
	A []string // all
	I []string // for ip_ table
	O []string // for op_ table
}

const emrURL = "root:root@tcp(192.168.9.5:3306)/emdata_emr_172"

var (
	sipTable map[string]struct{}
	sipts    = []string{
		"ip_patient_info",
		"ip_patient_info_two",
		"ip_patient_hospital_rel",
		"ip_medical_history_info",
		"ip_medical_history_status",
		"ip_doc_record",
		"ip_doc_record_item",
		"ip_lab_main_info",
		"ip_lab_detail_info",
		"ip_exam_info",
		"ip_exam_info_file",
		"ip_exam_ucg_result",
		"ip_index_tumour",
		"ip_index_tumour_drug",
		"ip_temperature_main",
		"ip_temperature_detail",
		"ip_nurse_main",
		"ip_nurse_detail",
		"ip_order_info",
		"ip_order_long",
		"ip_order_short",
		"ip_order_out",
		"ip_order_operation",
		"ip_patient_visit",
	}

	sopTable map[string]struct{}
	sopts    = []string{
		"op_medical_history_info",
		"op_medical_history_status",
		"op_exam_info",
		"op_exam_info_file",
		"op_exam_ucg_result",
		"op_lab_main_info",
		"op_lab_detail_info",
	}

	// TODO: need to put all db client to a pool
	emrDb *sql.DB

	// AllTables store all tables of the database
	allTables tables

	projIDs = []int{10013}
)

func arrToSet(arr []string) map[string]struct{} {
	r := make(map[string]struct{})
	for _, t := range arr {
		r[t] = struct{}{}
	}
	return r
}

func init() {
	// table array to set
	sipTable = arrToSet(sipts)
	sopTable = arrToSet(sopts)

	// init db pool
	var err error
	emrDb, err = sql.Open("mysql", emrURL)
	if err != nil {
		log.Fatalf("Connect to %s failed: %s", emrURL, err)
	}
	log.Printf("Connect to %s OK\n", emrURL)
}

func initTables(dbName string) {
	ts := fmt.Sprintf(`
	select table_name from information_schema.tables WHERE TABLE_SCHEMA = '%s' 
	AND (table_name LIKE 'ip_%%' OR table_name LIKE 'op_%%')
	`, dbName)
	records, err := getRecords(ts)
	if err != nil {
		log.Fatalln("Init table faield:", err)
	}
	allTables.A = append(allTables.A, records...)
	for _, tn := range records {
		if strings.HasPrefix(tn, "ip_") {
			if _, ok := sipTable[tn]; ok == false {
				allTables.I = append(allTables.I, tn)
			}
		} else {
			if _, ok := sopTable[tn]; ok == false {
				allTables.O = append(allTables.O, tn)
			}
		}
	}
	log.Println(allTables)
}

func getRecords(qs string) ([]string, error) {
	var results []string
	rows, err := emrDb.Query(qs)
	if err != nil {
		return results, err
	}
	var tn string
	for rows.Next() {
		err = rows.Scan(&tn)
		if err != nil {
			return results, err
		}
		results = append(results, tn)
	}
	if err = rows.Err(); err != nil {
		return results, err
	}
	return results, nil
}

const historySQL = `
select * 
from (
	select u.*,s.project_id,s.active,s.status,s.failed,s.max_run,s.remark 
	from (
		select 'ip' as ip_or_op,id,patient_id,client_status from ip_medical_history_info
		union select 'op' as ip_or_op,id,patient_id,client_status from op_medical_history_info
	) u
left join sync_status s on u.id=s.id) t %s
`

func query(qs string) ([]map[string]string, error) {
	var results []map[string]string
	rows, err := emrDb.Query(qs)
	if err != nil {
		return results, err
	}
	columns, err := rows.Columns()
	if err != nil {
		return results, err
	}

	values := make([]string, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return results, err
		}
		item := make(map[string]string)
		for i, col := range values {
			item[columns[i]] = col
		}
		results = append(results, item)
	}
	if err = rows.Err(); err != nil {
		return results, err
	}
	return results, nil
}

type medicalInfo struct {
	IOO         string `orm:"ip_or_op"`
	ID          string `orm:"id"`
	PatientID   string `orm:"patient_id"`
	ClientStatu string `orm:"client_statu"`
	ProjectID   string `orm:"project_id"`
	Active      string `orm:"active"`
	Status      string `orm:"status"`
	Failed      string `orm:"failed"`
	MaxRun      string `orm:"max_run"`
	Remark      string `orm:"remark"`
}

func (m *medicalInfo) valueScan() map[string]interface{} {
	fields := make(map[string]reflect.Value)
	v := reflect.ValueOf(m).Elem() // the struct variable
	for i := 0; i < v.NumField(); i++ {
		fieldInfo := v.Type().Field(i) // a reflect.StructField
		tag := fieldInfo.Tag           // a reflect.StructTag
		name := tag.Get("orm")
		if name == "" {
			name = strings.ToLower(fieldInfo.Name)
		}
		fields[name] = v.Field(i)
	}
	return nil
}

func getBatch(pid int, size int, ipop string) {
	qw := fmt.Sprintf(`where (status is null or status<1) 
	and (failed is null or failed<max_run) 
	and ip_or_op='%s' and (project_id is null or project_id=%d) 
	order by client_status asc, active asc LIMIT %d`, ipop, pid, size)
	qs := fmt.Sprintf(historySQL, qw)
	log.Println(qs)
	results, err := query(qs)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(results)
	s, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Println(err)
	}
	log.Println(string(s))
}
