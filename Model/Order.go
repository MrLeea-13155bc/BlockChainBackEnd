package Model

import (
	"fmt"
	"log"
	"main/Utils"
	"strconv"
	"sync"
)

func GetAllOrder(companyId int64) ([]Utils.Order, error) {
	var orderInfos []Utils.Order
	var orderInfo Utils.Order
	var seaTransportCompanyId int64
	template := `Select OrderId, StartDate,SeaTransportCompanyId, OrderStatus From Orders Where ClientCompanyId = ?`
	rows, err := Utils.DB().Query(template, companyId)
	if err != nil {
		log.Println("[GetAllOrder]数据库异常", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&orderInfo.OrderId, &orderInfo.StartDate, &seaTransportCompanyId, &orderInfo.Status)
		orderInfo.SeaTransCompanyName, _ = GetCompanyBasicInfo(seaTransportCompanyId)
		orderInfo.ClientCompanyName, _ = GetCompanyBasicInfo(companyId)
		orderInfos = append(orderInfos, orderInfo)
	}
	return orderInfos, nil
}

func RecordOrder(info Utils.OrderInfo) (int64, bool, error) {
	affair, err := Utils.DB().Begin()
	if err != nil {
		return 0, false, err
	}
	template := `Insert Into Orders Set ClientCompanyId = ?,StartDate = now() ,OrderStatus = ?`
	rows, err := affair.Exec(template, info.ClientCompanyId, "议价")
	if err != nil {
		log.Println("[RecordOrder]Orders出错了", err)
		affair.Rollback()
		return 0, false, err
	}
	info.OrderId, _ = rows.LastInsertId()
	wg := sync.WaitGroup{}
	template = `Insert Into Cargo Set CargoName = ? , CargoModel = ? , CargoSize = ? , CargoNum = ? , Category = ? , Weight = ? `
	for _, item := range info.Cargos {
		wg.Add(1)
		rows, err = affair.Exec(template, item.CargoName, item.CargoModel, item.CargoSize, item.CargoNum, item.Category, item.CargoWeight)
		if err != nil {
			log.Println("[RecordOrder]Cargo出错了", err)
			affair.Rollback()
			return 0, false, nil
		}
		cargoId, _ := rows.LastInsertId()
		go func(cargoId int64) {
			template := `Insert Into Order_Cargo Set OrderId = ?,CargoId = ?`
			affair.Exec(template, info.OrderId, cargoId)
			wg.Done()
		}(cargoId)
	}
	template = `Insert Into Address Set Country = ?,City = ?,Address = ?`
	rows, err = affair.Exec(template, info.SendAddress.Country, info.SendAddress.City, info.SendAddress.Address)
	sendAddressId, _ := rows.LastInsertId()
	rows, err = affair.Exec(template, info.ReceiveAddress.Country, info.ReceiveAddress.City, info.ReceiveAddress.Address)
	receiveAddressId, _ := rows.LastInsertId()
	template = `Insert Into OrderInfo Set OrderId = ?,StartAddressId = ? ,EndAddressId = ? ,Phone= ?,Email = ?,Fax = ? , HopeReachDate = ? , INCOTERMS = ? , UnStackable = ? , Perishable =?,Dangerous = ? , Clearance = ? , Other = ?, deliveryDate = ?`
	_, err = affair.Exec(template, info.OrderId, sendAddressId, receiveAddressId, info.Phone, info.Email, info.Fax, info.HopeReachDate, info.Incoterms, info.UnStackable, info.Perishable, info.Dangerous, info.Clearance, info.Other, info.DeliveryDate)
	wg.Wait()
	affair.Commit()
	return info.OrderId, true, nil
}

func CheckOrderCompany(orderId, companyId int64) bool {
	template := `Select ClientCompanyId From Orders Where OrderId = ?`
	rows, err := Utils.DB().Query(template, orderId)
	if err != nil || !rows.Next() {
		log.Println("[CheckOrderCompany] make a mistake", err)
		return false
	}
	var id int64
	rows.Scan(&id)
	return id == companyId
}

func GetCompanyBargain(orderId, companyId int64) ([]Utils.Bargain, error) {
	template := `Select B.CompanyId,COALESCE(Price,0),COALESCE(isPass,-1) From 
	(Select * From Bargain Where OrderId = ?) A Right Join (
		Select TargetCompanyId as companyId From Relation Where CompanyId = ? And isDelete = 0
		UNION
		Select CompanyId as companyId From Relation Where TargetCompanyId = ? And isDelete = 0
	) B On A.CompanyId = B.companyId`
	rows, err := Utils.DB().Query(template, orderId, companyId, companyId)
	if err != nil {
		return nil, err
	}
	var bargains []Utils.Bargain
	var bargain Utils.Bargain
	for rows.Next() {
		rows.Scan(&bargain.CompanyId, &bargain.Price, &bargain.Status)
		bargains = append(bargains, bargain)
	}
	return bargains, nil
}

func GetOrderInfo(orderId int64) (Utils.OrderInfo, error) {
	var info Utils.OrderInfo
	template := `Select OrderId, StartAddressId, EndAddressId, Phone, Email, Fax, HopeReachDate,deliveryDate,INCOTERMS, UnStackable, Perishable, Dangerous, Clearance, Other From OrderInfo Where OrderId = ? limit 1`
	rows, err := Utils.DB().Query(template, orderId)
	if err != nil {
		return info, err
	}
	defer rows.Close()
	if !rows.Next() {
		return info, fmt.Errorf("订单未查询到")
	}
	var startAddressId, endAddressId int64
	err = rows.Scan(&info.OrderId, &startAddressId, &endAddressId, &info.Phone, &info.Email, &info.Fax, &info.HopeReachDate, &info.DeliveryDate, &info.Incoterms, &info.UnStackable, &info.Perishable, &info.Dangerous, &info.Clearance, &info.Other)
	if err != nil {
		log.Println("[GetOrderInfo] make a mistake ", err)
		return info, err
	}
	defer rows.Close()
	template = `Select Country, City, Address From Address Where AddressId = ?`
	rows, err = Utils.DB().Query(template, startAddressId)
	if err != nil || !rows.Next() {
		log.Println("[GetOrderInfo] make a mistake ", err)
		return info, err
	}
	defer rows.Close()
	rows.Scan(&info.SendAddress.Country, &info.SendAddress.City, &info.SendAddress.Address)
	defer rows.Close()
	rows, err = Utils.DB().Query(template, endAddressId)
	if err != nil || !rows.Next() {
		log.Println("[GetOrderInfo] make a mistake ", err)
		return info, err
	}
	defer rows.Close()
	rows.Scan(&info.ReceiveAddress.Country, &info.ReceiveAddress.City, &info.ReceiveAddress.Address)
	defer rows.Close()
	template = `Select CargoId From Order_Cargo Where OrderId = ?`
	rows, err = Utils.DB().Query(template, orderId)
	if err != nil {
		log.Println("[GetOrderInfo] make a mistake ", err)
		return info, err
	}
	defer rows.Close()
	template = `Select CargoName, CargoModel, CargoNum, Category, Weight, CargoSize From Cargo Where CargoId = ? limit 1`
	var cargoId int64
	var cargo Utils.Cargo
	for rows.Next() {
		rows.Scan(&cargoId)
		rows1, err := Utils.DB().Query(template, cargoId)
		if err != nil || !rows1.Next() {
			log.Println("[GetOrderInfo] make a mistake ", err)
			continue
		}
		rows1.Scan(&cargo.CargoName, &cargo.CargoModel, &cargo.CargoNum, &cargo.Category, &cargo.CargoWeight, &cargo.CargoSize)
		cargo.CargoId = cargoId
		info.Cargos = append(info.Cargos, cargo)
		rows1.Close()
	}
	return info, nil
}

func CheckOrderCanBargain(orderId int64) bool {
	template := `Select OrderStatus  From Orders Where OrderId = ? limit 1`
	var status string
	rows, err := Utils.DB().Query(template, orderId)
	if err != nil || !rows.Next() {
		log.Println("[CheckOrderCanBargain] make a mistake ", err)
		return false
	}
	defer rows.Close()
	rows.Scan(&status)
	return status == "议价"
}

func ReplyBargain(bargain Utils.ReplyBargain, companyId int64) bool {
	template := `Update Bargain Set isPass = 1 , Price = ? ,ReplyTime = now() Where OrderId = ? And CompanyId = ?`
	result, err := Utils.DB().Exec(template, bargain.Bargain, bargain.OrderId, companyId)
	if err != nil {
		log.Println("[ReplyBargain]", err)
		return false
	}
	num, err := result.RowsAffected()
	return num == 1
}

func AskFroBargain(companyId, orderId int64) bool {
	template := `Insert Into Bargain Set ReplyTime = now() , OrderId = ? , CompanyId = ?`
	result, err := Utils.DB().Exec(template, orderId, companyId)
	if err != nil {
		log.Println("[AskForBargain] make a mistake ", err)
		return false
	}
	num, _ := result.RowsAffected()
	return num == 1
}

func UpdateOrderAgent(info Utils.OrderCompany) bool {
	template := `Update Orders Set SeaTransportCompanyId = ? , OrderStatus = '配置中' Where OrderId = ? limit 1`
	result, err := Utils.DB().Exec(template, info.SeaCompanyId, info.OrderId)
	if err != nil {
		log.Println("[UpdateOrderAgent] Make a mistake ", err)
		return false
	}
	line, _ := result.RowsAffected()
	return line == 1
}

func GetOrderClientId(orderId int64) (int64, error) {
	template := `Select ClientCompanyId From Orders Where OrderId = ?`
	rows, err := Utils.DB().Query(template, orderId)
	if err != nil || !rows.Next() {
		log.Println("[GetOrderClientId] make a mistake ", err)
		return -1, err
	}
	defer rows.Close()
	var companyId int64
	rows.Scan(&companyId)
	return companyId, nil
}

func CheckBargainSent(orderId, companyId int64) bool {
	template := `Select isPass From Bargain Where OrderId = ? And CompanyId = ?`
	rows, err := Utils.DB().Query(template, orderId, companyId)
	if err != nil {
		log.Println("[CheckBargainSent] make a mistake ", err)
		return false
	}
	defer rows.Close()
	return rows.Next()
}

func CheckOrderStatus(orderId int64, status string) bool {
	template := `Select OrderStatus From Orders Where OrderId = ?`
	rows, err := Utils.DB().Query(template, orderId)
	if err != nil {
		log.Println("[CheckOrderStatus] make a mistake ", err)
		return false
	}
	defer rows.Close()
	if !rows.Next() {
		return false
	}
	var curStatus string
	rows.Scan(&curStatus)
	return curStatus == status
}

func NoticeChoose(info Utils.OrderCompany, companyId int64) {
	companyName, _ := GetCompanyBasicInfo(companyId)
	text := `恭喜您，` + companyName + ` ( id :` + strconv.FormatInt(companyId, 10) +
		` ) 对于您对订单 ( id :` + strconv.FormatInt(info.OrderId, 10) + ` )使用了您的报价`
	text1 := text + strconv.FormatInt(info.SeaBargain, 10)
	SendMessageTo(0, text1, info.SeaCompanyId, 0)
}
