var searchInput = document.getElementById('search-input');
var suggestions = document.getElementById('suggestions');
var tags = [];
// 从 index.txt 读取所有标签
fetch('/preview/search/index.txt')
    .then(response => response.text())
    .then(text => {
        tags = text.split('\n').filter(Boolean);
    });
// 显示匹配的标签作为联想词
function showSuggestions(value) {
    var filteredTags = tags.filter(tag => tag.toLowerCase().includes(value.toLowerCase()));
    suggestions.innerHTML = '';
    filteredTags.forEach(tag => {
        var li = document.createElement('li');
        li.textContent = tag;
        li.onclick = () => {
            window.location.href = '/preview/tags/' + encodeURIComponent(tag) + '/';
        };
        suggestions.appendChild(li);
    });
}
// 输入框事件监听
searchInput.addEventListener('input', function() {
    var value = searchInput.value.trim();
     if (value) {
        showSuggestions(value);
    } else {
        suggestions.innerHTML = '';
    }
});
