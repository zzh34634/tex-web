package stat

import (
    "fmt"
    "time"
    "strconv"
    "github.com/labstack/echo"
    mid "github.com/yellia1989/tex-web/backend/middleware"
    "github.com/yellia1989/tex-web/backend/common"
    "github.com/yellia1989/tex-web/backend/api/gm"
)

type _remainlog struct {
    Statymd string `json:"statymd"`
    zoneid uint32
    Zonename string `json:"zone_name"`
    Zoneopenday uint32 `json:"zone_openday"`
    Days []float32 `json:"days"`
    Newadd float32 `json:"newadd"`
    Ltv7 float32 `json:"ltv7"`
    Ltv30 float32 `json:"ltv30"`
    RechargeRolenum float32 `json:"recharge_rolenum"`
    RechargeRate string `json:"recharge_rate"`
    RechargeMoney float32 `json:"recharge_money"`
    RechargePer string `json:"recharge_per"`
    RechargePer2 string `json:"recharge_per2"`
    RechargeR float32 `json:"recharge_r"`
}

type statkey struct {
    statymd string
    zoneid uint32
}

type statval struct {
    rolenum float32
    money float32
    money7 float32
    money30 float32
    active []float32
    r float32
    rechargerolenum float32
}

func RemainList(c echo.Context) error {
    ctx := c.(*mid.Context)
    zoneid := ctx.QueryParam("zoneid")
    page, _ := strconv.Atoi(ctx.QueryParam("page"))
    limit, _ := strconv.Atoi(ctx.QueryParam("limit"))
    startTime := ctx.QueryParam("startTime")
    endTime := ctx.QueryParam("endTime")
    datatype := ctx.QueryParam("datatype")

    now,_ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
    if startTime == "" {
        startTime = now.Add(-7*24*time.Hour).Format("2006-01-02 15:04:05")
    }
    if endTime == "" {
        endTime = now.Format("2006-01-02 15:04:05")
    }

    endTime2 := endTime
    t2,_ := time.Parse("2006-01-02 15:04:05", endTime2)
    t3 := t2.Add(24*90*time.Hour)
    if t3.After(now) {
        endTime2 = now.Format("2006-01-02 15:04:05")
    }

    mzone := gm.ZoneMap()

    db := common.GetStatDb()
    if db == nil {
        return ctx.SendError(-1, "连接数据库失败")
    }

    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    _, err = tx.Exec("USE db_stat")
    if err != nil {
        return err
    }

    sql := "SELECT statymd,zoneid"
    for i := 0;i < 90; i++ {
        sql += fmt.Sprintf(",sum(day_%d) as sum_day_%d",i+1,i+1)
    }
    sql += " FROM t_remain"
    sql2 := sql
    sql += " WHERE datatype="+datatype+" AND statymd between '"+startTime+"' AND '"+endTime+"'" 
    sql2 += " WHERE datatype="+datatype+" AND statymd between '"+startTime+"' AND '"+endTime2+"'" 
    if zoneid != "" {
        sql += " AND zoneid in ("+zoneid+")"
        sql2 += " AND zoneid in ("+zoneid+")"
    }
    sql += " GROUP BY statymd, zoneid ORDER BY statymd desc, zoneid desc"
    sql2 += " GROUP BY statymd, zoneid ORDER BY statymd desc, zoneid desc"

    c.Logger().Error(sql)

    var total int
    err = tx.QueryRow("SELECT count(*) as total FROM ("+sql+") a").Scan(&total)
    if err != nil {
        return err
    }

    limitstart := strconv.Itoa((page-1)*limit)
    limitrow := strconv.Itoa(limit)
    sql2 += " LIMIT "+limitstart+","+limitrow

    c.Logger().Error(sql2)

    rows, err := tx.Query(sql2)
    if err != nil {
        return err
    }
    defer rows.Close()

    // 充值记录查询开始时间,因为分页关系老的充值记录不用查询
    rechargeStartTime := ""

    logs := make([]_remainlog, 0)
    mdays := make(map[statkey]*statval,0)
    act := make([]float32,90)
    for rows.Next() {
        var r _remainlog
        if err := rows.Scan(&r.Statymd, &r.zoneid, &act[0], &act[1], &act[2], &act[3], &act[4], &act[5], &act[6], &act[7], &act[8], &act[9], &act[10], &act[11], &act[12], &act[13], &act[14], &act[15], &act[16], &act[17], &act[18], &act[19], &act[20], &act[21], &act[22], &act[23], &act[24], &act[25], &act[26], &act[27], &act[28], &act[29], &act[30], &act[31], &act[32], &act[33], &act[34], &act[35], &act[36], &act[37], &act[38], &act[39], &act[40], &act[41], &act[42], &act[43], &act[44], &act[45], &act[46], &act[47], &act[48], &act[49], &act[50], &act[51], &act[52], &act[53], &act[54], &act[55], &act[56], &act[57], &act[58], &act[59], &act[60], &act[61], &act[62], &act[63], &act[64], &act[65], &act[66], &act[67], &act[68], &act[69], &act[70], &act[71], &act[72], &act[73], &act[74], &act[75], &act[76], &act[77], &act[78], &act[79], &act[80], &act[81], &act[82], &act[83], &act[84], &act[85], &act[86], &act[87], &act[88], &act[89]); err != nil {
            return err
        }
        v := statval{}
        v.active = make([]float32,90)
        copy(v.active, act)
        k := statkey{statymd:r.Statymd, zoneid:r.zoneid}
        mdays[k] = &v

        if v,ok := mzone[r.zoneid]; ok {
            r.Zonename = v.SZoneName
            r.Zoneopenday = uint32(now.Sub(time.Unix(int64(v.IPublishTime),0)).Hours()/24)
        }
        t,_ := time.Parse("2006-01-02", r.Statymd)
        if !t.After(t2) {
            r.Newadd = act[0]
            r.Days = append(r.Days,act[0])
            r.RechargeRate = "0.0%"
            r.RechargePer = "0.00"
            r.RechargePer2 = "0.00"
            logs = append(logs, r)
            rechargeStartTime = r.Statymd
        }
    }
    if err := rows.Err(); err != nil {
        return err
    }

    c.Logger().Error(mdays)

    sqlwhere := " WHERE rolecreatetimeymd >= '" + rechargeStartTime + "'"
    if zoneid != "" {
        sqlwhere += " AND zoneid in ("+zoneid+")"
    }

    // 总充值金额和玩家数量
    sql3 := "SELECT rolecreatetimeymd, zoneid, count( DISTINCT roleid ) AS rolenum, sum( money ) AS money FROM t_recharge"
    sql3 += sqlwhere
    sql3 += " GROUP BY rolecreatetimeymd, zoneid ORDER BY rolecreatetimeymd, zoneid"
    c.Logger().Error(sql3)
    rows2, err := tx.Query(sql3)
    if err != nil {
        return err
    }
    defer rows2.Close()
    for rows2.Next() {
        var k statkey
        var rolenum float32
        var money float32
        if err := rows2.Scan(&k.statymd, &k.zoneid, &rolenum, &money); err != nil {
            return err
        }
        if v,ok := mdays[k]; ok {
            v.rolenum = rolenum
            v.money = money
        }
    }
    // 30ltv
    sql3 = "SELECT rolecreatetimeymd, zoneid, floor(( unix_timestamp( statymd )- unix_timestamp( rolecreatetimeymd ))/ 86400 ) AS days, sum( money ) AS money FROM t_recharge"
    sql3 += sqlwhere
    sql3 += " GROUP BY rolecreatetimeymd, statymd, zoneid HAVING floor(( unix_timestamp( statymd )- unix_timestamp( rolecreatetimeymd ))/ 86400 ) <=30"
    c.Logger().Error(sql3)
    rows3, err := tx.Query(sql3)
    if err != nil {
        return err
    }
    defer rows3.Close()
    for rows3.Next() {
        var k statkey
        var days uint32
        var money float32
        if err := rows2.Scan(&k.statymd, &k.zoneid, &days, &money); err != nil {
            return err
        }

        if v,ok := mdays[k]; ok {
            if days <= 7 {
                v.money7 += money
            }
            if days <= 30 {
                v.money30 += money
            }
        }
    }
    // 大R数量
    sql3 = "SELECT rolecreatetimeymd, zoneid, sum( money ) AS money FROM t_recharge" 
    sql3 += sqlwhere
    sql3 += " GROUP BY rolecreatetimeymd, zoneid,roleid"
    c.Logger().Error(sql3)
    rows4, err := tx.Query(sql3)
    if err != nil {
        return err
    }
    defer rows4.Close()
    for rows4.Next() {
        var k statkey
        var money float32
        if err := rows2.Scan(&k.statymd, &k.zoneid, &money); err != nil {
            return err
        }

        if v,ok := mdays[k]; ok {
            v.rechargerolenum += 1
            if (money > 200) {
                v.r += 1
            }
        }
    }

    if err := tx.Commit(); err != nil {
        return err
    }

    for i,_ := range logs {
        rlog := &logs[i]
        if v,ok := mdays[statkey{statymd:rlog.Statymd,zoneid:rlog.zoneid}]; ok {
            rlog.Ltv7 = v.money7
            rlog.Ltv30 = v.money30
            rlog.RechargeRolenum = v.rechargerolenum
            rlog.RechargeMoney = v.money
            rlog.RechargeR = v.r
            if rlog.Newadd != 0 {
                rlog.RechargeRate = fmt.Sprintf("%.2f%%", rlog.RechargeRolenum/rlog.Newadd)
                rlog.RechargePer2 = fmt.Sprintf("%.2f", rlog.RechargeMoney/rlog.Newadd)
            }
            if rlog.RechargeRolenum != 0 {
                rlog.RechargePer = fmt.Sprintf("%.2f", rlog.RechargeMoney/rlog.RechargeRolenum)
            }
        }
        t,_ := time.Parse("2006-01-02", rlog.Statymd)
        for j := 1; j < 90; j++ {
            t2 := t.Add(time.Duration(j*24)*time.Hour)
            if !t2.Before(now) {
                break
            }
            day := t2.Format("2006-01-02")
            active := float32(0)

            if v, ok := mdays[statkey{statymd:day,zoneid:rlog.zoneid}]; ok {
                active = v.active[j]
            }

            rlog.Days = append(rlog.Days, active)
        }
    }

    return ctx.SendArray(logs, total)
}
