package main

import (
	"errors"
	dns "github.com/alibabacloud-go/alidns-20150109/v2/client"
	env "github.com/alibabacloud-go/darabonba-env/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	console "github.com/alibabacloud-go/tea-console/client"
	util "github.com/alibabacloud-go/tea-utils/service"
	tea "github.com/alibabacloud-go/tea/tea"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// getCurrentHostIP
// 获取当前主机外网IP
func getCurrentHostIP() (s string, _err error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://ipinfo.io", nil)
	resp, err := client.Do(req)
	if err != nil {
		return s, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			println(err)
		}
	}(resp.Body)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return s, err
	}
	dat := util.ToMap(body)
	ip := dat["ip"]
	if ip == nil {
		return s, errors.New("未获取到ip地址")
	}
	return ip.(string), nil
}

// Initialization
// 初始化公共请求参数
func Initialization(regionId *string) (_result *dns.Client, _err error) {
	config := &openapi.Config{}
	// 您的AccessKey ID
	accessKeyId := env.GetEnv(tea.String("ADR_ALIYUN_ACCESS_KEY_ID"))
	if tea.StringValue(accessKeyId) == "" {
		return nil, errors.New("请检查环境变量[ADR_ALIYUN_ACCESS_KEY_ID]")
	}
	config.AccessKeyId = accessKeyId
	// config.AccessKeyId = tea.String("xxx")
	// 您的AccessKey Secret
	accessKeySecret := env.GetEnv(tea.String("ADR_ALIYUN_ACCESS_KEY_SECRET"))
	if tea.StringValue(accessKeySecret) == "" {
		return nil, errors.New("请检查环境变量[ADR_ALIYUN_ACCESS_KEY_SECRET]")
	}
	config.AccessKeySecret = accessKeySecret
	// config.AccessKeySecret = tea.String("xxx")
	// 您的可用区ID
	config.RegionId = regionId
	_result = &dns.Client{}
	_result, _err = dns.NewClient(config)
	return _result, _err
}

// DescribeDomainRecords
// 获取主域名的所有解析记录列表
func DescribeDomainRecords(client *dns.Client, domainName *string, RR *string, recordType *string) (_result *dns.DescribeDomainRecordsResponse, _err error) {
	req := &dns.DescribeDomainRecordsRequest{}
	// 主域名
	req.DomainName = domainName
	// 主机记录
	req.RRKeyWord = RR
	// 解析记录类型
	req.Type = recordType
	_, tryErr := func() (_r *dns.DescribeDomainRecordsResponse, _e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		resp, _err := client.DescribeDomainRecords(req)
		if _err != nil {
			return _result, _err
		}
		console.Log(tea.String("获取主域名的所有解析记录列表"))
		console.Log(util.ToJSONString(tea.ToMap(resp)))
		_result = resp
		return _result, _err
	}()
	if tryErr != nil {
		var sdkError = &tea.SDKError{}
		if _t, ok := tryErr.(*tea.SDKError); ok {
			sdkError = _t
		} else {
			sdkError.Message = tea.String(tryErr.Error())
		}
		console.Log(sdkError.Message)
	}
	return _result, _err
}

// UpdateDomainRecord
// 修改解析记录
func UpdateDomainRecord(client *dns.Client, req *dns.UpdateDomainRecordRequest) (_err error) {
	tryErr := func() (_e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		resp, _err := client.UpdateDomainRecord(req)
		if _err != nil {
			return _err
		}
		console.Log(tea.String("修改解析记录完成"))
		console.Log(util.ToJSONString(tea.ToMap(resp)))
		return nil
	}()

	if tryErr != nil {
		var sdkError = &tea.SDKError{}
		if _t, ok := tryErr.(*tea.SDKError); ok {
			sdkError = _t
		} else {
			sdkError.Message = tea.String(tryErr.Error())
		}
		console.Log(sdkError.Message)
	}
	return _err
}

// handleUpdateDomainRecord
// 处理更新域名记录
func handleUpdateDomainRecord(args []*string) (_err error) {
	regionId := args[0]
	currentHostIP := args[1]
	domainName := args[2]
	RR := args[3]
	recordType := args[4]
	client, _err := Initialization(regionId)
	if _err != nil {
		return _err
	}

	resp, _err := DescribeDomainRecords(client, domainName, RR, recordType)
	if _err != nil {
		return _err
	}

	if tea.BoolValue(util.IsUnset(tea.ToMap(resp))) || tea.BoolValue(util.IsUnset(tea.ToMap(resp.Body.DomainRecords.Record[0]))) {
		console.Log(tea.String("错误参数"))
		return _err
	}

	record := resp.Body.DomainRecords.Record[0]
	// 记录ID
	recordId := record.RecordId
	// 记录值
	recordsValue := record.Value
	if !tea.BoolValue(util.EqualString(currentHostIP, recordsValue)) {
		// 修改解析记录
		req := &dns.UpdateDomainRecordRequest{}
		// 主机记录
		req.RR = RR
		// 记录ID
		req.RecordId = recordId
		// 将主机记录值改为当前主机IP
		req.Value = currentHostIP
		// 解析记录类型
		req.Type = recordType
		_err = UpdateDomainRecord(client, req)
		if _err != nil {
			return _err
		}
	}
	return _err
}

// main
// 主流程
func _main(currentHostIP string) {
	var args []string
	// region id
	args = append(args, "cn-hangzhou")
	// currentHostIP
	args = append(args, currentHostIP)
	// domainName
	domainName := env.GetEnv(tea.String("ADR_DOMAIN_NAME"))
	// domainName = tea.String("xxx.com")
	if tea.StringValue(domainName) == "" {
		panic("请检查环境变量[ADR_DOMAIN_NAME]")
	}
	args = append(args, tea.StringValue(domainName))
	// RR 主机记录
	args = append(args, "@")
	// recordType 解析记录类型
	args = append(args, "A")
	err := handleUpdateDomainRecord(tea.StringSlice(args))
	if err != nil {
		panic(err)
	}
}

// main
// 主函数
func main() {
	cachedCurrentHostIP := ""
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		if cachedCurrentHostIP != "" {
			<-ticker.C
		}
		currentHostIP, err := getCurrentHostIP()
		if err != nil {
			println(err)
			continue
		}
		if cachedCurrentHostIP != currentHostIP {
			console.Log(tea.String("当前主机IP地址发生变化，开始更新域名解析"))
			console.Log(tea.String(cachedCurrentHostIP + " -> " + currentHostIP))
			_main(currentHostIP)
			cachedCurrentHostIP = currentHostIP
		}
	}
}
