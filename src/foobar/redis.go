package main

// import (
//     "crypto/tls"
//     "context"
//     "errors"
//     "github.com/go-redis/redis/v8"
//     "strconv"
//     "time"
// )

// func initRedis() {

//     verificationTimeoutMinutes, err := strconv.Atoi(getOptionalEnv("VERIFICATION_TIMEOUT_MINUTES", "5"))
//     if err != nil {
//         panic("VERIFICATION_TIMEOUT_MINUTES Not Int Value: " + err.Error())
//     }

//     verificationTimeout = time.Duration(verificationTimeoutMinutes)*time.Minute

//     baseContext = context.Background()
//     connectToRedis()

//     debugLog("Sucessfully established redis connection")
// }

// func connectToRedis() {

//     var redisOptions *redis.Options
//     if getOptionalEnv("REDIS_URL", "") != "" {
//         var err error
//         if redisOptions, err = redis.ParseURL(getRequiredEnv("REDIS_URL")); err != nil {
//             logError(err, nil)
//             panic(err)
//         }
//     } else {
//         redisURL := getRequiredEnv("REDIS_HOST") + ":" + getOptionalEnv("REDIS_PORT", "6379")
//         redisPassword := getOptionalEnv("REDIS_PASSWORD", "")
//         redisOptions = &redis.Options{
//                             Addr: redisURL,
//                             Password: redisPassword,
//                         }
//     }

//     // TODO- possible need for heroku
//     if getOptionalEnv("USE_TLS_CONFIG", "false") == "true" {
//         redisOptions.TLSConfig = &tls.Config{
//             InsecureSkipVerify: true,
//         }
//     }

//     redisClient = redis.NewClient(redisOptions)

//     var err error
//     for tries := 0; tries == 0 || err != nil; tries++ {
//         _, err = redisClient.Ping(baseContext).Result()

//         if err != nil {
//             if tries > 2 {
//                 logError(errors.New("Error: Could not connect to Redis"), nil)
//                 panic(err)
//             }
//             debugLog("Error: Could not connect to Redis -- trying again in 3 seconds")
//             time.Sleep(time.Duration(3)*time.Second)
//         }
//     }
// }
