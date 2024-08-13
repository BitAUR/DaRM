document.addEventListener("DOMContentLoaded", function() {
    // 获取页面中的所有图片元素
    const images = document.querySelectorAll("img");

    images.forEach(function(img) {
        // 为图片元素添加居中样式
        img.style.display = "block";
        img.style.marginLeft = "auto";
        img.style.marginRight = "auto";
    });
});
