package static

import (
	"fmt"
	"strings"

	"github.com/open-source/game/chess.git/pkg/xlog"
)

// --使用UNLINK删除，区别于del的是这个是异步执行的
// --这条指令要版本大于4.0.0 小于4.0.0就使用del
// redis.replicate_commands() 兼容redis版本
// var RedisBatchDeleteScript = redis.NewScript(`local cursor = 0
// local keyNum = 0
// repeat
// 	local res = redis.call("scan",KEYS[1],"MATCH",KEYS[2])
// 	if(res ~= nil and #res>=0) then
// 		cursor = tonumber(res[1])
// 		local ks = res[2]
// 		if(ks ~= nil and #ks>0) then
// 			redis.replicate_commands()
// 			for i=1,#ks,1 do
// 				local key = tostring(ks[i])
// 				redis.call("del",key)
// 			end
// 			keyNum = keyNum + #ks
// 		end
// 	end
// until( cursor <= 0 )
// return keyNum`)

/*

var RedisHGetHGet = redis.NewScript(`
local pkgKey = redis.call('HGET',KEYS[1],ARGV[1])

if (pkgKey ~= nil and pkgKey) then
	return redis.call('HGET',KEYS[2],pkgKey)
end

return nil
`)
//
var RedisHMGetHMGet = redis.NewScript(`
local res = {}
local pkeys = redis.call("HMGET",KEYS[1],unpack(ARGV))
if (pkeys ~= nil) then
	for k, v in pairs(pkeys) do
		if (v ~= nil) then
			local r = tostring(v)
				if ( r ~= "false") then
				table.insert(res, r)
				end
		end
	end
end
return redis.call("HMGET",KEYS[2],unpack(res))
`)
//
// var RedisSUnionHMGet = redis.NewScript(`
// local res = redis.call("SUNION",unpack(KEYS))
// if (res ~= nil and #res > 0) then
// 	return redis.call("HMGET",ARGV[1],unpack(res))
// end
// return nil
// `)

// var RedisSAddMulti = redis.NewScript(`
// local cvt = function(input, delimiter)
//     input = tostring(input)
//     delimiter = tostring(delimiter)
//     if (delimiter=='') then return false end
//     local pos,arr = 0, {}
//     for st,sp in function() return string.find(input, delimiter, pos, true) end do
//         table.insert(arr, string.sub(input, pos, st - 1))
//         pos = sp + 1
//     end
//     table.insert(arr, string.sub(input, pos))
//     return arr
// end
// for k, v in pairs(ARGV) do
// 	local res = cvt(v, "#")
// 	redis.call("SADD",KEYS[k],unpack(res))
// end
// return true
// `)
*/

func RedisSwitchObjectsToStrings(res []interface{}, err error) ([]string, error) {
	result := make([]string, 0)
	for _, o := range res {
		if o == nil {
			continue
		}
		s, ok := o.(string)
		if !ok {
			xlog.Logger().Errorln("错误: 不能转换一个底层非string类型")
			continue
		}
		result = append(result, s)
	}
	return result, err
}

func RedisSwitchObjectsToString(res []interface{}) string {
	result := make([]string, 0)
	for _, o := range res {
		if o == nil {
			continue
		}
		result = append(result, fmt.Sprintf("%+v", o))
	}
	return strings.Join(result, "#")
}
