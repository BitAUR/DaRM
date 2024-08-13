package main

// 定义网页模板

const BaseTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>DaRM</title>
	<link href="./res/full.min.css" rel="stylesheet" type="text/css" />
	<script src="./res/tailwind.js"></script>
    <link rel="icon" href="./res/logo.png">
</head>
<body class="text-gray-800">
<div class="navbar bg-base-100">
  <div class="navbar-start">
    <div class="dropdown">
      <div tabindex="0" role="button" class="btn btn-ghost lg:hidden">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h8m-8 6h16" /></svg>
      </div>
	  <ul tabindex="0" class="menu menu-sm dropdown-content mt-3 z-[1] p-2 shadow bg-base-100 rounded-box w-52">
		<li>
		<a>文章</a>
	  	<ul class="p-2">
		 <li><a href="./new">新建</a></li>
         <li><a href="./article">列表</a></li>
	   </ul>
		</li>
		<li><a target="_blank" rel="noopener" href="./preview/">预览</a></li>
		<li>
		<a>部署</a>
		<ul class="p-2">
		   <li><a href="./deploy">简介</a></li>
		   <li><a href="./ftp">FTP</a></li>
		   <li><a href="./github">GitHub</a></li>
		 </ul>
	</li>
    <li><a href="./settings">设置</a></li>
	</ul>
    </div>
    <a href="/"class="btn btn-ghost btn-wide text-xl">DaRM</a>
  </div>
  <div class="navbar-center hidden lg:flex">
    <ul class="menu menu-horizontal px-1">
    <li>
        <details>
  	  		<summary>文章</summary>
	 		<ul class="p-2">
				<li><a href="./new">新建</a></li>
                <li><a href="./article">列表</a></li>
	  		</ul>
        </details>
    </li>
	<li><a target="_blank" rel="noopener" href="./preview/">预览</a></li>
	<li>
		<details>
			<summary>部署</summary>
		   <ul class="p-2">
			  <li><a href="./deploy">简介</a></li>
			  <li><a href="./ftp">FTP</a></li>
			  <li><a href="./github">GitHub</a></li>
			</ul>
		</details>
	</li>
    <li><a href="./settings">设置</a></li>
    </ul>
  </div>
  <div class="navbar-end">
  <form action="/generate" method="post">
  <input type="hidden" id="redirectUrl" name="redirectUrl" value="">
  <button id="generateButton" onclick="generateBlog()" type="submit" class="btn btn-wide">生成</button>
  </form>
  </div>
</div>
{{.Content}}
</body>
<script>
document.addEventListener('DOMContentLoaded', (event) => {
    document.getElementById('redirectUrl').value = window.location.pathname;
});
</script>
<script>
document.addEventListener('DOMContentLoaded', (event) => {
    const urlParams = new URLSearchParams(window.location.search);
    const generateSuccess = urlParams.get('generateSuccess');
    
    if (generateSuccess === 'true') {
        var saveButton = document.getElementById('generateButton');
        generateButton.style.backgroundColor = 'green';
        generateButton.style.color = '#fff';
        generateButton.textContent = '成功';

        setTimeout(function() {
            generateButton.style.backgroundColor = '';
            generateButton.style.color = '';
            generateButton.textContent = '生成';
        }, 3000);
    }
            if (generateSuccess === 'false') {
        var saveButton = document.getElementById('generateButton');
        generateButton.style.backgroundColor = 'red';
        generateButton.style.color = '#fff';
        generateButton.textContent = '失败';

        setTimeout(function() {
            generateButton.style.backgroundColor = '';
            generateButton.style.color = '';
            generateButton.textContent = '生成';
        }, 3000);
    }
});
</script>
</html>
`

const loginForm = `
<!DOCTYPE html>
<html>
<head>
    <title>Login - DaRM</title>
	<link href="./res/full.min.css" rel="stylesheet" type="text/css" />
	<script src="./res/tailwind.js"></script>
        <link rel="icon" href="./res/logo.png">

</head>
<body class="flex items-center justify-center h-screen">
    <div class="form-control w-full max-w-xs">
		<div class="text-5xl font-bold text-center" style="user-select:none;">DaRM</div>
		<form method="post" action="/login" class="bg-base-100 p-8 box-border rounded">
			<div class="form-control mb-4">
				<input type="text" name="username" placeholder="username" class="input input-bordered w-full max-w-xs" required>
			</div>
			<div class="form-control mb-4">
				<input type="password" name="password" placeholder="password" class="input input-bordered w-full max-w-xs" required>
			</div>
			<div class="form-control mt-6" id="login-button-container">
				<button type="submit" class="btn btn-wide primary">登录</button>
			</div>
		</form>
		{{if .LoginFailed}}
		<script>
			var btnContainer = document.getElementById('login-button-container');
			btnContainer.innerHTML = '<button style="background-color:red;color:#fff" type="submit" class="btn btn-wide primary">账号或密码错误</button>';
			
			setTimeout(function() {
				btnContainer.innerHTML = '<button type="submit" class="btn btn-wide primary">Login</button>';
			}, 3000);
		</script>
		{{end}}
    </div>
</body>
</html>
`
const HomePageContent = `
<div class="p-8 centered" style="user-select:none;">
    <h1  class="text-8xl font-bold">DaRM</h1>
	<a></a>
</div>
`

const articlesTemplate = `
<div class="overflow-x-auto centered">
  <table class="table">
    <thead>
      <tr>
        <th>标题</th>
        <th>URI</th>
        <th>分类</th>
		<th>日期</th>
		<th></th>
		<th></th>
      </tr>
    </thead>
    <tbody>
      <!-- row 1 -->
	  {{range .}}
      <tr class="hover">
        <td>{{.Title}}</td>
        <td>{{.URI}}</td>
        <td>{{.Category}}</td>
		<td>{{.Date}}</td>
		<td><a href="/edit?title={{.Title}}">编辑</a></td>
		<td><a href="#" onclick="confirmDelete('{{.Title}}');">删除</a></td>
      </tr>
	  {{end}}
    </tbody>
  </table>
</div>
<script>
function confirmDelete(title) {
    var confirmed = confirm("确定要删除 " + title + " 么?");
    if (confirmed) {
        // 如果用户确认，重定向到删除URL，并附带确认参数
        window.location.href = '/delete?title=' + encodeURIComponent(title) + '&confirm=true';
    }
}
</script>
`
const newArticle = `
<div class="centered">
<form id="newArticleForm" method="post" action="/new" class="p8">
	<div>
	<div class="p-8" style="user-select:none;">
    <h1  class="text-3xl font-bold text-center" >新建文章</h1>
	<a></a>
	</div>
	<div class="form-control mb-4">
		<input type="text" id="title" name="title" placeholder="标题" class="input input-bordered w-full max-w-xs" required>
		</div>
		<div class="form-control mb-4">
		<input type="text" id="description" name="description" placeholder="描述" class="input input-bordered w-full max-w-xs" required>
		</div>
		<div class="form-control mb-4">
		<input type="text" id="category" name="category" placeholder="分类" class="input input-bordered w-full max-w-xs" required>
		</div>
		<div class="form-control mb-4">
		<input type="text" id="tags" name="tags" placeholder="使用逗号分隔多个标签" class="input input-bordered w-full max-w-xs" required>
		</div>
		<div class="form-control mb-4">
		<input type="date" id="date" name="date" placeholder="时间" class="input input-bordered w-full max-w-xs" required>
		</div>
		<div class="form-control mb-4">
		<input type="text" id="uri" name="uri" placeholder="URI" class="input input-bordered w-full max-w-xs" required>
		</div>
		</div>
	<div class="form-control mt-6" id="login-button-container">
		<button type="submit" class="btn btn-wide primary">创建文章</button>
	</div>
</form>
</div>
`
const settingsTemplate = `
<div class="centered">
<form id="settingsForm" method="post" action="/settings" class="p8">
    <div class="p-8" style="user-select:none;">
        <h1  class="text-3xl font-bold text-center" >设置</h1>
    </div>
    <div class="form-control mb-4">
        <input type="text" id="username" name="username" placeholder="账号" value="{{.UserName}}" class="input input-bordered w-full max-w-xs" required>
    </div>
	<div class="form-control mb-4">
        <input type="password" id="password" name="password" placeholder="密码" value="{{.Password}}" class="input input-bordered w-full max-w-xs" required>
    </div>
    <div class="form-control mb-4">
        <input type="text" id="blogtitle" name="blogtitle" placeholder="博客名称" value="{{.BlogTitle}}" class="input input-bordered w-full max-w-xs" required>
    </div>
    <div class="form-control mb-4">
        <input type="text" id="blogdescription" name="blogdescription" placeholder="博客简介" value="{{.Description}}" class="input input-bordered w-full max-w-xs" required>
    </div>
    <div class="form-control mb-4">
        <input type="text" id="blogtags" name="blogtags" placeholder="博客标签" value="{{.Tags}}" class="input input-bordered w-full max-w-xs" required>
    </div>
    <div class="form-control mb-4">
        <input type="text" id="bloguri" name="bloguri" placeholder="博客地址" value="{{.URI}}" class="input input-bordered w-full max-w-xs" required>
    </div>
    <div class="form-control mb-4">
		<input type="text" id="blogauthor" name="blogauthor" placeholder="博客作者" value="{{.Author}}" class="input input-bordered w-full max-w-xs" required>
    </div>
    <div class="form-control mb-4">
        <input type="text" id="email" name="email" placeholder="电子邮箱" value="{{.Email}}" class="input input-bordered w-full max-w-xs" required>
    </div>
    <div class="form-control mb-4">
        <input type="text" id="commenturi" name="commenturi" placeholder="评论系统地址" value="{{.CommentURI}}" class="input input-bordered w-full max-w-xs" required>
    </div>
    <div class="form-control mt-6" id="save-button-container">
        <button type="submit" id="saveButton" class="btn btn-wide primary">保存</button>
    </div>
</form>
<script>
document.addEventListener('DOMContentLoaded', (event) => {
    const urlParams = new URLSearchParams(window.location.search);
    const saved = urlParams.get('saved');
    
    if (saved === 'true') {
        var saveButton = document.getElementById('saveButton');
        saveButton.style.backgroundColor = 'green';
        saveButton.style.color = '#fff';
        saveButton.textContent = '保存成功';

        setTimeout(function() {
            saveButton.style.backgroundColor = '';
            saveButton.style.color = '';
            saveButton.textContent = '保存';
        }, 3000);
    }
});
</script>
</div>

`

const DeployContent = `
<div class="p-8 centered" style="user-select:none;">
    <h1  class="text-4xl font-bold">Deploy</h1>
    <br>
	<a>该功能目前支持 FTP 和 Github 推送</a>
    <br>
    <a>核心逻辑是将 /public 生成的静态文件，推送到对应的服务。</a>
    <br>
    <a>如使用 Github，可配合 CloudFlare Pages 或 Vercel 自动部署。</a>
</div>
`

// FTP 页面的 HTML 模板
const ftpTemplate = `
    <div class="centered">
<form id="newArticleForm" method="post" action="/ftp" class="p8">
	<div class="p-8" style="user-select:none;">
    <h1  class="text-3xl font-bold text-center" >FTP</h1>
	<a></a>
	</div>
	<div class="form-control mb-4">
		<input type="text" id="server" name="server" placeholder="服务器" value={{.Server}} class="input input-bordered w-full max-w-xs" required>
		</div>
		<div class="form-control mb-4">
		<input type="text" id="port" name="port" placeholder="端口" value={{.Port}} class="input input-bordered w-full max-w-xs" required>
		</div>
		<div class="form-control mb-4">
		<input type="text" id="username" name="username" placeholder="账号" value={{.Username}} class="input input-bordered w-full max-w-xs" required>
		</div>
		<div class="form-control mb-4">
		<input type="password" id="password" name="password" placeholder="密码" value={{.Password}} class="input input-bordered w-full max-w-xs" required>
		</div>
		<div class="form-control mb-4">
        <input type="text" id="relpath" name="relpath" placeholder="相对路径" value="{{.RelPath}}" class="input input-bordered w-full max-w-xs" required>
		</div>
        <div class="form-control">
            <label class="label cursor-pointer">
            <a>是否推送：</a>
                <input type="checkbox" id="push" name="push"  {{if .Push}}checked{{end}} class="checkbox" />
            </label>
        </div>
	<div class="form-control mt-6" id="login-button-container">
		<button id="PushButton" type="submit" class="btn btn-wide primary">{{if .Push}}保存并推送{{else}}保存{{end}}</button>
	</div>
</form>
</div>

<script>
document.addEventListener('DOMContentLoaded', function() {
    var checkbox = document.getElementById('push');
    var button = document.querySelector('button[id="PushButton"]');

    // 定义一个函数来根据复选框的状态更新按钮的标题
    function updateButtonLabel() {
        if (checkbox.checked) {
            button.textContent = '保存并推送';
        } else {
            button.textContent = '保存';
        }
    }

    // 在页面加载时设置正确的按钮标题
    updateButtonLabel();

    // 为复选框添加事件监听器，当其状态变化时更新按钮的标题
    checkbox.addEventListener('change', updateButtonLabel);
});
</script>
<script>
document.addEventListener('DOMContentLoaded', (event) => {
    const urlParams = new URLSearchParams(window.location.search);
    const success = urlParams.get('success');
    var checkbox = document.getElementById('push');

    if (success === 'true') {
        var PushButton = document.getElementById('PushButton');
        PushButton.style.backgroundColor = 'green';
        PushButton.style.color = '#fff';
        PushButton.textContent = '成功';

        setTimeout(function() {
            PushButton.style.backgroundColor = '';
            PushButton.style.color = '';
            if (checkbox.checked) {
                PushButton.textContent = '保存并推送';
                } else {
                PushButton.textContent = '保存';
                }
        }, 3000);
    }
    if (success == 'false'){
        var PushButton = document.getElementById('PushButton');
        PushButton.style.backgroundColor = 'red';
        PushButton.style.color = '#fff';
        PushButton.textContent = '失败';

        setTimeout(function() {
            PushButton.style.backgroundColor = '';
            PushButton.style.color = '';
            if (checkbox.checked) {
                PushButton.textContent = '保存并推送';
                } else {
                PushButton.textContent = '保存';
                }
        }, 3000);
    }
});
</script>
`

const githubTemplate = `
<div class="centered">
<div class="p-8" style="user-select:none;">
    <h1  class="text-3xl font-bold text-center" >GitHub</h1>
	<a></a>
	</div>
<form id="githubForm" method="post" action="/github" class="p8">
    <div class="form-control mb-4">
        <input type="text" id="repository" name="repository" placeholder="GitHub 存储库 URL" value="{{.Repository}}"  class="input input-bordered w-full max-w-xs" required>
    </div>
    <div class="form-control mb-4">
        <input type="text" id="branch" name="branch" placeholder="分支" value="{{.Branch}}"  class="input input-bordered w-full max-w-xs" required>
    </div>
    <div class="form-control mb-4">
        <input type="text" id="username" name="username" placeholder="账号" value="{{.Username}}"  class="input input-bordered w-full max-w-xs" required>
    </div>
    <div class="form-control mb-4">
    <input type="password" id="token" name="token" placeholder="Token" value="{{.Token}}"  class="input input-bordered w-full max-w-xs" required>
    </div>
    <div class="form-control mb-4">
        <input type="email" id="email" name="email" placeholder="邮箱" value="{{.Email}}"  class="input input-bordered w-full max-w-xs" required>
    </div>
    <div class="form-control">
    <label class="label cursor-pointer">
    <a>是否推送：</a>
        <input type="checkbox" id="push" name="push"  {{if .Push}}checked{{end}} class="checkbox" />
    </label>
</div>
    <div class="form-control mt-6">
    <button id="PushButton" type="submit" class="btn btn-wide primary">{{if .Push}}保存并推送{{else}}保存{{end}}</button>
    </div>
</form>
</div>

<script>
document.addEventListener('DOMContentLoaded', function() {
    var checkbox = document.getElementById('push');
    var button = document.querySelector('button[id="PushButton"]');

    // 定义一个函数来根据复选框的状态更新按钮的标题
    function updateButtonLabel() {
        if (checkbox.checked) {
            button.textContent = '保存并推送';
        } else {
            button.textContent = '保存';
        }
    }

    // 在页面加载时设置正确的按钮标题
    updateButtonLabel();

    // 为复选框添加事件监听器，当其状态变化时更新按钮的标题
    checkbox.addEventListener('change', updateButtonLabel);
});
</script>
<script>
document.addEventListener('DOMContentLoaded', (event) => {
    const urlParams = new URLSearchParams(window.location.search);
    const success = urlParams.get('success');
    var checkbox = document.getElementById('push');

    if (success === 'true') {
        var PushButton = document.getElementById('PushButton');
        PushButton.style.backgroundColor = 'green';
        PushButton.style.color = '#fff';
        PushButton.textContent = '成功';

        setTimeout(function() {
            PushButton.style.backgroundColor = '';
            PushButton.style.color = '';
            if (checkbox.checked) {
                PushButton.textContent = '保存并推送';
                } else {
                PushButton.textContent = '保存';
                }
        }, 3000);
    }
    if (success == 'false'){
        var PushButton = document.getElementById('PushButton');
        PushButton.style.backgroundColor = 'red';
        PushButton.style.color = '#fff';
        PushButton.textContent = '失败';

        setTimeout(function() {
            PushButton.style.backgroundColor = '';
            PushButton.style.color = '';
            if (checkbox.checked) {
                PushButton.textContent = '保存并推送';
                } else {
                PushButton.textContent = '保存';
                }
        }, 3000);
    }
});
</script>

`
const editTemplate = `
<div style="display: flex; flex-direction: column; justify-content: center; align-items: center; height: 80vh;">
<form method="post" style="width: 100%; height: 100%; display: flex; flex-direction: column; justify-content: center; align-items: center;">
<div style="width: 70%; height: 80%; display: flex;">
            <div style="text-align: left; width: 60%;">
                <textarea id="markdown" name="content" style="resize: none; width: 100%; height: 100%;" class="textarea h-24 textarea-bordered">{{.Content}}</textarea>
            </div>
            <div class="divider divider-horizontal"></div> 
            <div id="preview" class="textarea h-24 textarea-bordered" style="overflow: auto; margin-left: auto; text-align: left; width: 60%; height: 100%;">
            </div>
        </div>
        <div style="width: 100%; text-align: center;margin-top: 20px">
            <button id="saveButton" type="submit" class="btn btn-wide">保存</button>
        </div>
    </form>
</div>
<script src="/res/marked.min.js"></script>
<script>
document.addEventListener('DOMContentLoaded', function() {
    function updatePreview() {
        var markdownText = document.getElementById('markdown').value;
        document.getElementById('preview').innerHTML = marked.parse(markdownText);
    }

    document.getElementById('markdown').addEventListener('input', updatePreview);
    updatePreview(); // 初始化时立即更新预览
});
</script>
<script>
document.addEventListener('DOMContentLoaded', (event) => {
    const urlParams = new URLSearchParams(window.location.search);
    const save = urlParams.get('save');

    if (save === 'success') {
        var saveButton = document.getElementById('saveButton');
        saveButton.style.backgroundColor = 'green';
        saveButton.style.color = '#fff';
        saveButton.textContent = '成功';

        setTimeout(function() {
            saveButton.style.backgroundColor = '';
            saveButton.style.color = '';
            saveButton.textContent = '保存';
        }, 3000);
    }
});
</script>
`
