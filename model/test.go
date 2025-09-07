package models

// import (
//     "fmt"
//     "golang.org/x/crypto/bcrypt"
// )

// // TestPasswordCheck 独立测试密码验证功能
// func TestPasswordCheck() {
//     fmt.Println("=== 密码验证功能测试 ===")
    
//     // 测试用例1: 使用已知的正确密码和哈希
//     testPassword := "1145141919" // 替换为实际测试密码
//     testHash := "$2a$10$mvoca8uUbBsC.GFeqODYZeLua9ULxNu5shZjKBJZPlIA0Eodi9q5G" // 替换为实际哈希
    
//     fmt.Printf("测试密码: '%s'\n", testPassword)
//     fmt.Printf("测试哈希: '%s'\n", testHash)
    
//     // 直接使用 bcrypt 验证
//     err := bcrypt.CompareHashAndPassword([]byte(testHash), []byte(testPassword))
//     if err != nil {
//         fmt.Printf("直接 bcrypt 验证失败: %v\n", err)
//     } else {
//         fmt.Println("直接 bcrypt 验证成功!")
//     }
    
//     // 使用 User 结构体验证
//     testUser := User{
//         Password: testHash,
//     }
    
//     err = testUser.CheckPassword(testPassword)
//     if err != nil {
//         fmt.Printf("User.CheckPassword 验证失败: %v\n", err)
//     } else {
//         fmt.Println("User.CheckPassword 验证成功!")
//     }
    
//     // 测试错误密码
//     fmt.Println("\n=== 测试错误密码 ===")
//     err = testUser.CheckPassword("wrong_password")
//     if err != nil {
//         fmt.Printf("错误密码验证失败 (符合预期): %v\n", err)
//     } else {
//         fmt.Println("错误密码验证成功 (不符合预期)!")
//     }
// }