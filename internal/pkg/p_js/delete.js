(function() {
    const className = '{delete_class_name}';
    // 1. 参数校验：确保class名传入有效
    if (!className) {
        console.error("错误：class名必须非空");
    }

    // 2. 构造CSS类选择器（. + 类名）
    const selector = `.${className}`;

    // 3. 查找所有匹配元素
    const elements = document.querySelectorAll(selector);
    const deleteCount = elements.length;

    // 4. 遍历批量删除元素
    elements.forEach(element => {
        // 兼容旧浏览器的安全判断（确保元素可被删除）
        if (element && element.parentNode) {
            element.remove();
        }
    });

    // 5. 返回执行结果（包含详细信息，方便Golang端解析）
   console.log( `批量删除完成，已处理匹配元素`, deleteCount, className)
})
