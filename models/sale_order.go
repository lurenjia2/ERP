package models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"goERP/utils"

	"github.com/astaxie/beego/orm"
)

// SaleOrder 产品分类
type SaleOrder struct {
	ID         int64            `orm:"column(id);pk;auto" json:"id"`         //主键
	CreateUser *User            `orm:"rel(fk);null" json:"-"`                //创建者
	UpdateUser *User            `orm:"rel(fk);null" json:"-"`                //最后更新者
	CreateDate time.Time        `orm:"auto_now_add;type(datetime)" json:"-"` //创建时间
	UpdateDate time.Time        `orm:"auto_now;type(datetime)" json:"-"`     //最后更新时间
	FormAction string           `orm:"-" form:"FormAction"`                  //非数据库字段，用于表示记录的增加，修改
	Name       string           `orm:"unique" json:"name"`                   //订单号
	Partner    *Partner         `orm:"rel(fk)"`                              //客户
	SalesMan   *User            `orm:"rel(fk)"`                              //业务员
	Company    *Company         `orm:"rel(fk)"`                              //公司
	Country    *AddressCountry  `orm:"rel(fk);null" json:"country"`          //国家
	Province   *AddressProvince `orm:"rel(fk);null" json:"province"`         //省份
	City       *AddressCity     `orm:"rel(fk);null" json:"city"`             //城市
	District   *AddressDistrict `orm:"rel(fk);null" json:"district"`         //区县
	Street     string           `orm:"default(\"\")" json:"street"`          //街道
	OrderLine  []*SaleOrderLine `orm:"reverse(many)"`                        //订单明细
	State      *SaleOrderState  `orm:"rel(fk)"`                              //订单状态
}

func init() {
	orm.RegisterModel(new(SaleOrder))
}

// AddSaleOrder insert a new SaleOrder into database and returns
// last inserted ID on success.
func AddSaleOrder(obj *SaleOrder) (id int64, err error) {
	o := orm.NewOrm()

	id, err = o.Insert(obj)
	return id, err
}

// GetSaleOrderByID retrieves SaleOrder by ID. Returns error if
// ID doesn't exist
func GetSaleOrderByID(id int64) (obj *SaleOrder, err error) {
	o := orm.NewOrm()
	obj = &SaleOrder{ID: id}
	if err = o.Read(obj); err == nil {
		return obj, nil
	}
	return nil, err
}

// GetAllSaleOrder retrieves all SaleOrder matches certain condition. Returns empty list if
// no records exist
func GetAllSaleOrder(query map[string]interface{}, exclude map[string]interface{}, condMap map[string]map[string]interface{}, fields []string, sortby []string, order []string, offset int64, limit int64) (utils.Paginator, []SaleOrder, error) {
	var (
		objArrs   []SaleOrder
		paginator utils.Paginator
		num       int64
		err       error
	)
	if limit == 0 {
		limit = 20
	}
	o := orm.NewOrm()
	qs := o.QueryTable(new(SaleOrder))
	qs = qs.RelatedSel()

	//cond k=v cond必须放到Filter和Exclude前面
	cond := orm.NewCondition()
	if _, ok := condMap["and"]; ok {
		andMap := condMap["and"]
		for k, v := range andMap {
			k = strings.Replace(k, ".", "__", -1)
			cond = cond.And(k, v)
		}
	}
	if _, ok := condMap["or"]; ok {
		orMap := condMap["or"]
		for k, v := range orMap {
			k = strings.Replace(k, ".", "__", -1)
			cond = cond.Or(k, v)
		}
	}
	qs = qs.SetCond(cond)
	// query k=v
	for k, v := range query {
		// rewrite dot-notation to Object__Attribute
		k = strings.Replace(k, ".", "__", -1)
		qs = qs.Filter(k, v)
	}
	//exclude k=v
	for k, v := range exclude {
		// rewrite dot-notation to Object__Attribute
		k = strings.Replace(k, ".", "__", -1)
		qs = qs.Exclude(k, v)
	}

	// order by:
	var sortFields []string
	if len(sortby) != 0 {
		if len(sortby) == len(order) {
			// 1) for each sort field, there is an associated order
			for i, v := range sortby {
				orderby := ""
				if order[i] == "desc" {
					orderby = "-" + strings.Replace(v, ".", "__", -1)
				} else if order[i] == "asc" {
					orderby =  strings.Replace(v, ".", "__", -1)
				} else {
					return paginator, nil, errors.New("Error: Invalid order. Must be either [asc|desc]")
				}
				sortFields = append(sortFields, orderby)
			}
			qs = qs.OrderBy(sortFields...)
		} else if len(sortby) != len(order) && len(order) == 1 {
			// 2) there is exactly one order, all the sorted fields will be sorted by this order
			for _, v := range sortby {
				orderby := ""
				if order[0] == "desc" {
					orderby = "-" + strings.Replace(v, ".", "__", -1)
				} else if order[0] == "asc" {
					orderby =  strings.Replace(v, ".", "__", -1)
				} else {
					return paginator, nil, errors.New("Error: Invalid order. Must be either [asc|desc]")
				}
				sortFields = append(sortFields, orderby)
			}
		} else if len(sortby) != len(order) && len(order) != 1 {
			return paginator, nil, errors.New("Error: 'sortby', 'order' sizes mismatch or 'order' size is not 1")
		}
	} else {
		if len(order) != 0 {
			return paginator, nil, errors.New("Error: unused 'order' fields")
		}
	}

	qs = qs.OrderBy(sortFields...)
	if cnt, err := qs.Count(); err == nil {
		paginator = utils.GenPaginator(limit, offset, cnt)
	}
	if num, err = qs.Limit(limit, offset).All(&objArrs, fields...); err == nil {
		paginator.CurrentPageSize = num
	}
	return paginator, objArrs, err
}

// UpdateSaleOrderByID updates SaleOrder by ID and returns error if
// the record to be updated doesn't exist
func UpdateSaleOrderByID(m *SaleOrder) (err error) {
	o := orm.NewOrm()
	v := SaleOrder{ID: m.ID}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Update(m); err == nil {
			fmt.Println("Number of records updated in database:", num)
		}
	}
	return
}

// GetSaleOrderByName retrieves SaleOrder by Name. Returns error if
// Name doesn't exist
func GetSaleOrderByName(name string) (obj *SaleOrder, err error) {
	o := orm.NewOrm()
	obj = &SaleOrder{Name: name}
	if err = o.Read(obj); err == nil {
		return obj, nil
	}
	return nil, err
}

// DeleteSaleOrder deletes SaleOrder by ID and returns error if
// the record to be deleted doesn't exist
func DeleteSaleOrder(id int64) (err error) {
	o := orm.NewOrm()
	v := SaleOrder{ID: id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Delete(&SaleOrder{ID: id}); err == nil {
			fmt.Println("Number of records deleted in database:", num)
		}
	}
	return
}
