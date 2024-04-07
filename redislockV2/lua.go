package redislockv2

const (
	LauCheckAndDelete = `
		if(redis.call('get',KEYS[1])==ARGV[1]) then
			return redis.call('del',KEYS[1])
		else
			return 0
		end
	`

	LauCheckThenExpire = `
    	if redis.call('get', KEYS[1]) == ARGV[1]
    	then 
      		return redis.call('expire', KEYS[1], ARGV[2]) 
    	else 
      		return 0 
  	  	end
  	`
)
