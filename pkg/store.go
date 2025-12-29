// Simplified logic for Hour 4
func CheckIdempotency(redisClient *redis.Client, key string) (bool, error) {
    // SETNX (Set if Not eXists) returns true if the key was created
    return redisClient.SetNX(ctx, key, "processing", 24*time.Hour).Result()
}