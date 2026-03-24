const PC_IP = window.location.hostname;
const backendHost = `http://${PC_IP}:8080`;
function renderVideos(videos) {
    const listElement = document.getElementById('video-list');
    listElement.innerHTML = '';

    if (!videos || videos.length === 0) {
        listElement.innerHTML = '<p>没有找到视频 ┐(′～`)┌</p>';
        return;
    }

    videos.forEach((video) => {
        // 1. 标题处理：如果后端没处理后缀，前端这里去掉
        const cleanTitle = video.title.replace(/\.[^/.]+$/, "");

        // 2. 💡 重要修改：使用 file_hash 而不是文件名
        // 如果后端返回的字段名是 file_hash，请确保一致
        const fileID = video.filehash || video.id;
        console.log(video);
        // 3. 海报逻辑
        const posterBaseApi = `${backendHost}/api/poster/`;
        const defaultPoster = '/images/default-cover.png';
        const posterUrl = (video.poster && video.poster.trim() !== "")
            ? `${posterBaseApi}${encodeURIComponent(video.poster)}`
            : defaultPoster;

        // 4. 创建视频卡片 HTML
        const card = document.createElement('div');
        card.className = 'video-card';
        card.style = `
            border: 1px solid #eee; 
            border-radius: 12px; 
            margin-bottom: 12px; 
            padding: 12px; 
            cursor: pointer;
            display: flex;
            align-items: center;
            background: #fff;
            transition: all 0.3s ease;
            box-shadow: 0 4px 6px rgba(0,0,0,0.02);
        `;

        // 悬停动效
        card.onmouseover = () => {
            card.style.transform = 'translateY(-2px)';
            card.style.boxShadow = '0 8px 15px rgba(0,0,0,0.08)';
        };
        card.onmouseout = () => {
            card.style.transform = 'translateY(0)';
            card.style.boxShadow = '0 4px 6px rgba(0,0,0,0.02)';
        };

        card.innerHTML = `
            <div class="cover-box" style="width: 130px; height: 80px; margin-right: 16px; flex-shrink: 0; overflow: hidden; border-radius: 8px; background: #f5f5f5;">
                <img src="${posterUrl}" 
                     alt="${cleanTitle}"
                     onerror="this.src='${defaultPoster}';" 
                     style="width: 100%; height: 100%; object-fit: cover;">
            </div>
            <div style="flex-grow: 1; min-width: 0;">
                <div style="font-size: 15px; font-weight: 600; color: #222; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;" title="${cleanTitle}">
                    ${cleanTitle}
                </div>
                <div style="font-size: 12px; color: #999; margin-top: 6px; display: flex; align-items: center;">
                    <span style="background: #eef2ff; color: #6366f1; padding: 2px 6px; border-radius: 4px; margin-right: 8px;">HLS</span>
                    点击播放
                </div>
            </div>
        `;

        // 5. 💡 点击事件：存储 Hash 而不是长文件名
        card.onclick = () => {
            // 存储 Hash 值，player.html 将通过这个 Hash 去找 m3u8 文件夹
            sessionStorage.setItem('play_file', fileID);
            sessionStorage.setItem('play_title', cleanTitle);
            window.location.href = 'player.html';
        };

        listElement.appendChild(card);
    });
}

// 搜索逻辑
async function handSearch() {
    const input = document.getElementById('search-bar');
    const keyword = input.value.trim();
    if(!keyword) return loadVideos(); // 空搜索则加载全部

    const listElement = document.getElementById('video-list');
    listElement.innerHTML = '<div class="loading">搜索中...</div>';

    try {
        const url = `${backendHost}/api/search?title=${encodeURIComponent(keyword)}`;
        const response = await fetch(url);
        const videos = await response.json();
        renderVideos(videos);
    } catch (err) {
        listElement.innerHTML = `<p style="color:red">搜索失败: ${err.message}</p>`;
    }
}

// 初始加载
async function loadVideos() {
    try {
        const response = await fetch(`${backendHost}/api/play`);
        const videos = await response.json();
        renderVideos(videos);
    } catch (err) {
        document.getElementById('video-list').innerHTML = '<p style="color:red">连接后端失败，请检查服务是否启动</p>';
    }
}

window.onload = () => {
    loadVideos();
    const searchInput = document.getElementById('search-bar');
    if (searchInput) {
        searchInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') handSearch();
        });
    }
};