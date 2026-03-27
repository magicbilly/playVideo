// 完全没有完成
async function login() {
    // 1. 修正方法名：是 getElementById (单数)，不是 getElementsById
    const name = document.getElementById("username").value;
    const password = document.getElementById("password").value;

    // 2. 修正 URL：前后端分离时，建议只在 Body 中传参，不要在 URL 后面拼接
    const url = "http://localhost:8080/api/login";

    try {
        // 3. 使用 await 等待响应
        const response = await fetch(url, {
            method: "POST",
            headers: {
                "Content-Type": "application/x-www-form-urlencoded",
            },
            // 4. 关键：添加 credentials 选项，否则跨域时 Cookie 不会被保存
            credentials: 'include',
            body: new URLSearchParams({
                "name": name,
                "password": password
            })
        });
        // 5. 处理跳转逻辑
        if (response.ok) {
            alert("登录成功！");
            // 跳转到前端 3000 端口的首页
            window.location.href = "/index.html";
        } else {
            const errorText = await response.text();
            alert("登录失败: " + errorText);
        }
    } catch (error) {
        console.error("请求失败:", error);
    }
}