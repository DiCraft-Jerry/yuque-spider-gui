<script>
  import { onDestroy, onMount } from 'svelte';
  import { fade } from 'svelte/transition';
  import {
    AddTask,
    RemoveTask,
    StartTask,
    CancelTask,
    GetAllTasks,
    StartAllPendingTasks,
    ClearCompletedTasks,
    SelectDirectory,
    GetDefaultConfig,
    ValidateURL,
    EventsOn,
    WAILS_BACKEND_ERR,
    wailsBackendUserMessage
  } from './appApi.js';

  /** @param {unknown} err @param {string} prefix */
  function formatAppError(err, prefix) {
    if (err && typeof err === 'object' && 'message' in err && err.message === WAILS_BACKEND_ERR) {
      return wailsBackendUserMessage();
    }
    return prefix + (err != null ? String(err) : '');
  }

  let tasks = [];

  let newTask = {
    url: '',
    cookie: '',
    outputPath: ''
  };

  let defaultOutputPath = '';

  let batchInput = '';
  let showBatchModal = false;
  let showHelpTip = false;
  let theme = 'light';

  let config = {
    delayMin: 1,
    delayMax: 4,
    timeout: 30,
    imageTimeout: 60,
    maxRetries: 3,
    concurrentDownloads: 1,
    downloadMode: 'lake',
    failOnImageError: false
  };

  $: stats = {
    total: tasks.length,
    pending: tasks.filter(t => t.status === 'pending').length,
    running: tasks.filter(t => t.status === 'running').length,
    completed: tasks.filter(t => t.status === 'completed').length,
    failed: tasks.filter(t => t.status === 'failed').length
  };

  let errorMessage = '';
  let successMessage = '';
  let hintTimer;

  function clearHintTimer() {
    if (hintTimer) {
      clearTimeout(hintTimer);
      hintTimer = undefined;
    }
  }

  function showSuccess(message, duration = 2500) {
    clearHintTimer();
    errorMessage = '';
    successMessage = message;
    hintTimer = setTimeout(() => {
      successMessage = '';
      hintTimer = undefined;
    }, duration);
  }

  function showError(message, duration = 4000) {
    clearHintTimer();
    successMessage = '';
    errorMessage = message;
    hintTimer = setTimeout(() => {
      errorMessage = '';
      hintTimer = undefined;
    }, duration);
  }

  onMount(async () => {
    const defaultConfig = await GetDefaultConfig();
    config = defaultConfig;

    await loadTasks();

    EventsOn('tasks:update', (taskList) => {
      tasks = taskList;
      hydrateDefaultOutput(taskList);
    });

    EventsOn('task:update', (task) => {
      const index = tasks.findIndex(t => t.id === task.id);
      if (index !== -1) {
        tasks[index] = task;
        tasks = [...tasks];
        hydrateDefaultOutput(tasks);
      }
    });
  });

  onDestroy(() => {
    clearHintTimer();
  });

  function hydrateDefaultOutput(taskList) {
    if (defaultOutputPath) return;

    const withPath = taskList?.find((t) => t.outputPath);
    if (withPath) {
      defaultOutputPath = withPath.outputPath;
      if (!newTask.outputPath) {
        newTask.outputPath = defaultOutputPath;
      }
    }
  }

  async function loadTasks() {
    try {
      const allTasks = await GetAllTasks();
      tasks = allTasks;
      hydrateDefaultOutput(allTasks);
    } catch (err) {
      console.error('加载任务失败:', err);
    }
  }

  async function selectOutputDir() {
    try {
      const dir = await SelectDirectory();
      if (dir) {
        defaultOutputPath = dir;
        newTask.outputPath = dir;
        showSuccess('已更新默认输出目录');
      }
    } catch (err) {
      console.error('选择目录失败:', err);
      showError('选择目录失败');
    }
  }

  async function addTask() {
    clearHintTimer();
    errorMessage = '';
    successMessage = '';

    const isValid = await ValidateURL(newTask.url);
    if (!isValid) {
      showError('请输入有效的语雀 URL');
      return;
    }

    const targetOutputPath = newTask.outputPath || defaultOutputPath;
    if (!targetOutputPath) {
      showError('请选择输出目录');
      return;
    }

    try {
      await AddTask(newTask.url, newTask.cookie, targetOutputPath, config);
      showSuccess('任务添加成功');

      defaultOutputPath = targetOutputPath;
      newTask.outputPath = targetOutputPath;
      newTask.url = '';
      newTask.cookie = '';
    } catch (err) {
      showError('添加任务失败: ' + err);
    }
  }

  async function removeTask(taskId) {
    try {
      await RemoveTask(taskId);
    } catch (err) {
      showError('删除任务失败: ' + err);
    }
  }

  async function startTask(taskId) {
    try {
      await StartTask(taskId);
    } catch (err) {
      showError('启动任务失败: ' + err);
    }
  }

  async function cancelTask(taskId) {
    try {
      await CancelTask(taskId);
    } catch (err) {
      showError('取消任务失败: ' + err);
    }
  }

  async function startAllPending() {
    try {
      await StartAllPendingTasks();
      showSuccess('已启动所有待处理任务');
    } catch (err) {
      showError('启动任务失败: ' + err);
    }
  }

  async function clearCompleted() {
    try {
      await ClearCompletedTasks();
      showSuccess('已清除完成或失败的任务');
    } catch (err) {
      showError('清除任务失败: ' + err);
    }
  }

  function parseBatchInput() {
    const lines = batchInput.trim().split('\n');
    const result = [];

    for (const line of lines) {
      const trimmed = line.trim();
      if (!trimmed) continue;

      const parts = trimmed.split(',');
      const url = parts[0].trim();
      const cookie = parts.length > 1 ? parts[1].trim() : '';

      if (url) {
        result.push({ url, cookie });
      }
    }

    return result;
  }

  async function importBatch() {
    clearHintTimer();
    errorMessage = '';
    successMessage = '';

    const targetOutputPath = newTask.outputPath || defaultOutputPath;
    if (!targetOutputPath) {
      showError('请先选择输出目录');
      return;
    }

    const batchTasks = parseBatchInput();

    if (batchTasks.length === 0) {
      showError('没有有效的任务');
      return;
    }

    let successCount = 0;
    for (const task of batchTasks) {
      try {
        await AddTask(task.url, task.cookie, targetOutputPath, config);
        successCount++;
      } catch (err) {
        console.error('添加任务失败:', err);
      }
    }

    defaultOutputPath = targetOutputPath;
    newTask.outputPath = targetOutputPath;

    showSuccess(`成功添加 ${successCount}/${batchTasks.length} 个任务`);

    showBatchModal = false;
    batchInput = '';
  }

  function getStatusBadgeClass(status) {
    const map = {
      pending: 'badge badge-pending',
      running: 'badge badge-running',
      completed: 'badge badge-completed',
      failed: 'badge badge-failed',
      cancelled: 'badge badge-cancelled'
    };
    return map[status] || 'badge';
  }

  function getStatusText(status) {
    const map = {
      pending: '等待中',
      running: '下载中',
      completed: '已完成',
      failed: '失败',
      cancelled: '已取消'
    };
    return map[status] || status;
  }

  function formatDate(dateStr) {
    if (!dateStr) return '-';
    const date = new Date(dateStr);
    return date.toLocaleString('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  function toggleTheme() {
    theme = theme === 'light' ? 'dark' : 'light';
  }

  function toggleHelpTip() {
    if (showHelpTip) {
      showHelpTip = false;
      return;
    }
    showHelpTip = true;
  }

  /** @param {MouseEvent} event */
  function handleGlobalClick(event) {
    if (!showHelpTip) return;
    const target = event.target;
    if (!(target instanceof Element)) return;
    if (target.closest('.help-tooltip')) return;
    showHelpTip = false;
  }
</script>

<svelte:window on:click={handleGlobalClick} />

<main class={`admin-app ${theme === 'dark' ? 'theme-dark' : ''}`}>
  <header class="app-header">
    <div class="brand">
      <div class="brand-main">
        <svg class="brand-icon" viewBox="0 0 64 64" fill="none" aria-hidden="true">
          <defs>
            <linearGradient id="brandFrame" x1="12" y1="10" x2="52" y2="54" gradientUnits="userSpaceOnUse">
              <stop stop-color="#6366F1" />
              <stop offset="1" stop-color="#8B5CF6" />
            </linearGradient>
            <linearGradient id="brandArrow" x1="32" y1="20" x2="32" y2="47" gradientUnits="userSpaceOnUse">
              <stop stop-color="#22D3EE" />
              <stop offset="1" stop-color="#2563EB" />
            </linearGradient>
          </defs>
          <rect x="8" y="8" width="48" height="48" rx="14" fill="url(#brandFrame)" />
          <path d="M22 20C22 18.9 22.9 18 24 18H36L42 24V44C42 45.1 41.1 46 40 46H24C22.9 46 22 45.1 22 44V20Z" fill="white" fill-opacity="0.94" />
          <path d="M36 18V22.4C36 23.28 36.72 24 37.6 24H42" fill="#E9D5FF" />
          <path d="M32 26V38" stroke="url(#brandArrow)" stroke-width="3.2" stroke-linecap="round" />
          <path d="M27 34L32 39L37 34" stroke="url(#brandArrow)" stroke-width="3.2" stroke-linecap="round" stroke-linejoin="round" />
          <rect x="25.5" y="42" width="13" height="2.8" rx="1.4" fill="#14B8A6" />
          <circle cx="50" cy="14" r="5" fill="#10B981" stroke="white" stroke-width="1.5" />
        </svg>
        <div class="brand-title">语雀下载器</div>
      </div>
    </div>
    <div class="header-stats">
      <div class="metric">
        <span class="metric-value">{stats.total}</span>
        <span class="metric-label">总任务</span>
      </div>
      <div class="metric">
        <span class="metric-value">{stats.running}</span>
        <span class="metric-label">运行中</span>
      </div>
      <div class="metric">
        <span class="metric-value">{stats.pending}</span>
        <span class="metric-label">等待中</span>
      </div>
      <div class="metric">
        <span class="metric-value">{stats.completed}</span>
        <span class="metric-label">已完成</span>
      </div>
    </div>
    <div class="header-actions">
      <div class="help-tooltip">
        <button
          class={`help-bulb ${showHelpTip ? 'is-active' : ''}`}
          type="button"
          on:click|stopPropagation={toggleHelpTip}
          aria-label="使用提示"
          title="使用提示"
        >
          <svg viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path d="M9 17h6M10 20h4M8 10a4 4 0 1 1 8 0c0 1.7-.8 2.8-1.8 3.8-.6.6-1.2 1.2-1.2 2.2h-2c0-1-.6-1.6-1.2-2.2C8.8 12.8 8 11.7 8 10Z" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </button>
        <div class={`help-popover ${showHelpTip ? 'is-visible' : ''}`}>
          <div>1. 先选择输出目录（仅需一次）</div>
          <div>2. 粘贴知识库 URL 与 Cookie</div>
          <div>3. 推荐先用 `lake`，md格式在网络波动时，图片会超时</div>
        </div>
      </div>
      <button
        on:click={toggleTheme}
        class={`theme-switch ${theme === 'dark' ? 'is-dark' : ''}`}
        type="button"
        aria-label="切换主题"
        title={theme === 'light' ? '切换到夜间模式' : '切换到白天模式'}
      >
        <span class="switch-icon switch-sun" aria-hidden="true">
          <svg viewBox="0 0 24 24" fill="none">
            <circle cx="12" cy="12" r="4" stroke="currentColor" stroke-width="1.8" />
            <path d="M12 2.5V5M12 19V21.5M2.5 12H5M19 12H21.5M5.2 5.2L7 7M17 17L18.8 18.8M18.8 5.2L17 7M7 17L5.2 18.8" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/>
          </svg>
        </span>
        <span class="switch-icon switch-moon" aria-hidden="true">
          <svg viewBox="0 0 24 24" fill="none">
            <path d="M20 14.2A8.2 8.2 0 1 1 9.8 4c-.2.5-.3 1.1-.3 1.7a6.8 6.8 0 0 0 6.8 6.8c1 0 2-.2 2.7-.7Z" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"/>
          </svg>
        </span>
        <span class="switch-thumb" aria-hidden="true"></span>
      </button>
    </div>
  </header>

  <div class="floating-hints" aria-live="polite" aria-atomic="true">
    {#if errorMessage}
      <div class="hint-toast hint-error" transition:fade={{ duration: 180 }}>
        ❌ {errorMessage}
      </div>
    {/if}
    {#if successMessage}
      <div class="hint-toast hint-success" transition:fade={{ duration: 180 }}>
        ✅ {successMessage}
      </div>
    {/if}
  </div>

  <div class="app-shell">
    <aside class="app-sidebar">
      <section class="sidebar-block">
        <h3>输出目录</h3>
        <div class="output-selector">
          <div class="output-path" title={defaultOutputPath || '未选择目录'}>
            {defaultOutputPath || '未选择'}
          </div>
          <button class="btn btn-secondary" on:click={selectOutputDir}>选择目录</button>
        </div>
        <p class="sidebar-hint">每个知识库会在该目录下创建一个同名文件夹。</p>
      </section>

      <section class="sidebar-block">
        <h3>下载配置</h3>
        <div class="config-grid">
          <div class="config-item">
            <label>延迟范围 (秒)</label>
            <div class="config-range">
              <input type="number" bind:value={config.delayMin} min="0" max="10" />
              <span>-</span>
              <input type="number" bind:value={config.delayMax} min="1" max="20" />
            </div>
          </div>
          <div class="config-item">
            <label>超时 (秒)</label>
            <input type="number" bind:value={config.timeout} min="10" max="120" />
          </div>
          <div class="config-item">
            <label>图片超时 (秒)</label>
            <input type="number" bind:value={config.imageTimeout} min="10" max="300" />
          </div>
          <div class="config-item">
            <label>文档类型</label>
            <select bind:value={config.downloadMode}>
              <option value="md">md</option>
              <option value="lake">lake</option>
            </select>
          </div>
          <div class="config-item">
            <label>图片失败策略</label>
            <select bind:value={config.failOnImageError}>
              <option value={false}>继续下载文档</option>
              <option value={true}>文档下载失败</option>
            </select>
          </div>
        </div>
      </section>

    </aside>

    <section class="app-content">
      <div class="card">
        <h2 class="card-title">新建任务</h2>
        <div class="form-grid">
          <label class="form-label">知识库 URL</label>
          <input
            type="text"
            bind:value={newTask.url}
            placeholder="https://www.yuque.com/user/book"
          />

          <label class="form-label">Cookie (可选)</label>
          <input
            type="text"
            bind:value={newTask.cookie}
            placeholder="访问私有知识库时填写"
          />

          <label class="form-label">保存路径</label>
          <div class="path-row">
            <input
              type="text"
              bind:value={newTask.outputPath}
              placeholder="未选择"
              readonly
            />
            <button class="btn btn-secondary" on:click={selectOutputDir}>选择目录</button>
            <button class="btn btn-primary" on:click={addTask}>➕ 添加</button>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="card-header">
          <h2 class="card-title">任务列表</h2>
          <div class="card-actions">
            <div class="card-subtitle">共 {stats.total} 个任务</div>
            <button on:click={() => showBatchModal = true} class="btn btn-outline">批量导入</button>
            <button on:click={startAllPending} disabled={stats.pending === 0} class="btn btn-primary">开始全部</button>
            <button on:click={clearCompleted} disabled={stats.completed === 0 && stats.failed === 0} class="btn btn-secondary">清除完成</button>
          </div>
        </div>

        {#if tasks.length === 0}
          <div class="empty-state">
            <div class="empty-icon">📭</div>
            <p>暂无下载任务</p>
            <p class="empty-hint">添加新任务或批量导入开始下载</p>
          </div>
        {:else}
          {#each tasks as task (task.id)}
            <div class="task-card">
              <div class="task-header">
                <div class="task-info">
                  <div class="task-url" title={task.url}>{task.url}</div>
                  <div class={getStatusBadgeClass(task.status)}>{getStatusText(task.status)}</div>
                </div>
                <div class="task-actions">
                  {#if task.status === 'pending'}
                    <button on:click={() => startTask(task.id)} class="btn-icon" title="开始">▶️</button>
                  {:else if task.status === 'running'}
                    <button on:click={() => cancelTask(task.id)} class="btn-icon" title="取消">⏸️</button>
                  {/if}
                  {#if task.status !== 'running'}
                    <button on:click={() => removeTask(task.id)} class="btn-icon btn-danger" title="删除">🗑️</button>
                  {/if}
                </div>
              </div>

              {#if task.progress && task.progress.bookTitle}
                <div class="task-book-title">📖 {task.progress.bookTitle}</div>
              {/if}

              <div class="task-meta">
                <span class="meta-label">保存到:</span>
                <span class="meta-value" title={task.outputPath}>{task.outputPath}</span>
              </div>

              {#if task.status === 'running' || (task.progress && task.progress.totalDocs > 0)}
                <div class="task-progress">
                  {#if task.progress.currentDoc}
                    <div class="progress-doc">📄 {task.progress.currentDoc}</div>
                  {/if}
                  <div class="progress-bar-container">
                    <div class="progress-bar" style={`width: ${task.progress.percentage || 0}%`}></div>
                  </div>
                  <div class="progress-summary">
                    <span>{task.progress.finishedDocs || 0} / {task.progress.totalDocs || 0} 文档</span>
                    <span>{Math.round(task.progress.percentage || 0)}%</span>
                  </div>
                </div>
              {/if}

              {#if task.error}
                <div class="task-error">⚠️ {task.error}</div>
              {/if}

              <div class="task-footer">
                <span>创建: {formatDate(task.createdAt)}</span>
                {#if task.completedAt}
                  <span>完成: {formatDate(task.completedAt)}</span>
                {/if}
              </div>
            </div>
          {/each}
        {/if}
      </div>
    </section>
  </div>

  {#if showBatchModal}
    <div class="modal-overlay" on:click={() => showBatchModal = false}>
      <div class="modal-content" on:click|stopPropagation>
        <div class="modal-header">
          <h3>批量导入任务</h3>
          <button class="modal-close" on:click={() => showBatchModal = false}>×</button>
        </div>
        <div class="modal-body">
          <p class="modal-hint">
            每行一个任务，格式: URL,Cookie (Cookie 可选)<br />
            例如:<br />
            https://www.yuque.com/user/book1<br />
            https://www.yuque.com/user/book2,yuque_session=xxx
          </p>
          <textarea
            bind:value={batchInput}
            placeholder="粘贴任务列表..."
            rows="10"
            class="batch-textarea"
          ></textarea>
        </div>
        <div class="modal-footer">
          <button on:click={() => showBatchModal = false} class="btn btn-secondary">取消</button>
          <button on:click={importBatch} class="btn btn-primary">导入</button>
        </div>
      </div>
    </div>
  {/if}
</main>

<style>
  :global(body) {
    margin: 0;
    padding: 0;
    font-family: 'Inter', 'PingFang SC', 'Microsoft YaHei', sans-serif;
    background: #f3f4f6;
    color: #1f2937;
  }

  .admin-app {
    --bg-page: #f3f4f6;
    --bg-panel: #ffffff;
    --bg-sidebar: #f8fafc;
    --bg-sidebar-card: #ffffff;
    --sidebar-text-main: #1f2937;
    --sidebar-text-sub: #64748b;
    --sidebar-border: #dbe2ea;
    --sidebar-input-bg: #f8fafc;
    --sidebar-input-border: #d0d7e2;
    --text-main: #1f2937;
    --text-sub: #6b7280;
    --line: #e5e7eb;
  }

  .admin-app.theme-dark {
    --bg-page: #0b1220;
    --bg-panel: #111827;
    --bg-sidebar: linear-gradient(180deg, #0b1220 0%, #111827 100%);
    --bg-sidebar-card: rgba(17, 24, 39, 0.72);
    --sidebar-text-main: #e5e7eb;
    --sidebar-text-sub: #9ca3af;
    --sidebar-border: rgba(255, 255, 255, 0.08);
    --sidebar-input-bg: rgba(2, 6, 23, 0.5);
    --sidebar-input-border: rgba(148, 163, 184, 0.2);
    --text-main: #e5e7eb;
    --text-sub: #9ca3af;
    --line: #233043;
  }

  .admin-app {
    display: flex;
    flex-direction: column;
    height: 100vh;
    background: var(--bg-page);
    color: var(--text-main);
  }

  .app-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 28px;
    background: var(--bg-panel);
    border-bottom: 1px solid var(--line);
    box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
  }

  .brand-title {
    font-size: 1.4rem;
    font-weight: 700;
  }

  .brand-main {
    display: inline-flex;
    align-items: center;
    gap: 10px;
  }

  .brand-icon {
    width: 34px;
    height: 34px;
    flex: 0 0 auto;
    filter: drop-shadow(0 2px 4px rgba(15, 23, 42, 0.25));
  }

  .brand-subtitle {
    font-size: 0.8rem;
    color: #6b7280;
    margin-top: 2px;
  }

  .header-stats {
    display: flex;
    gap: 20px;
  }

  .metric {
    display: flex;
    flex-direction: column;
    align-items: center;
  }

  .metric-value {
    font-size: 1.2rem;
    font-weight: 700;
  }

  .metric-label {
    font-size: 0.75rem;
    color: var(--text-sub);
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .theme-switch {
    position: relative;
    width: 70px;
    height: 34px;
    border-radius: 999px;
    border: 1px solid var(--line);
    background: linear-gradient(90deg, #f8fafc 0%, #e2e8f0 100%);
    display: inline-flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 8px;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .theme-switch:hover {
    border-color: #6366f1;
  }

  .theme-switch.is-dark {
    background: linear-gradient(90deg, #1e293b 0%, #0f172a 100%);
  }

  .switch-icon {
    width: 14px;
    height: 14px;
    display: inline-flex;
    color: #64748b;
    z-index: 1;
  }

  .switch-icon svg {
    width: 14px;
    height: 14px;
  }

  .theme-switch .switch-sun {
    color: #f59e0b;
  }

  .theme-switch .switch-moon {
    color: #64748b;
  }

  .theme-switch.is-dark .switch-sun {
    color: #94a3b8;
  }

  .theme-switch.is-dark .switch-moon {
    color: #cbd5e1;
  }

  .switch-thumb {
    position: absolute;
    top: 3px;
    left: 3px;
    width: 26px;
    height: 26px;
    border-radius: 50%;
    background: #ffffff;
    box-shadow: 0 1px 3px rgba(15, 23, 42, 0.35);
    transition: transform 0.2s ease;
  }

  .theme-switch.is-dark .switch-thumb {
    transform: translateX(36px);
    background: #cbd5e1;
  }

  .help-tooltip {
    position: relative;
  }

  .help-bulb {
    width: 36px;
    height: 36px;
    border-radius: 999px;
    border: 1px solid var(--line);
    background: #ffffff;
    color: #6b7280;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .help-bulb svg {
    width: 18px;
    height: 18px;
  }

  .help-bulb:hover {
    border-color: #f59e0b;
    color: #f59e0b;
  }

  .help-bulb.is-active {
    background: #fde68a;
    border-color: #f59e0b;
    color: #b45309;
    box-shadow: 0 0 0 2px rgba(245, 158, 11, 0.28), 0 0 16px rgba(245, 158, 11, 0.4);
  }

  .theme-dark .help-bulb {
    background: #1f2937;
    color: #94a3b8;
    border-color: #334155;
  }

  .theme-dark .help-bulb.is-active {
    background: #fbbf24;
    color: #7c2d12;
    border-color: #f59e0b;
    box-shadow: 0 0 0 2px rgba(245, 158, 11, 0.24), 0 0 16px rgba(251, 191, 36, 0.42);
  }

  .help-popover {
    position: absolute;
    top: calc(100% + 8px);
    right: 0;
    min-width: 300px;
    background: var(--bg-panel);
    border: 1px solid var(--line);
    border-radius: 10px;
    box-shadow: 0 8px 24px rgba(15, 23, 42, 0.16);
    padding: 10px 12px;
    font-size: 0.82rem;
    line-height: 1.55;
    color: var(--text-main);
    text-align: left;
    opacity: 0;
    pointer-events: none;
    transform: translateY(-4px);
    transition: all 0.16s ease;
    z-index: 20;
  }

  .help-popover.is-visible {
    opacity: 1;
    pointer-events: auto;
    transform: translateY(0);
  }

  .app-shell {
    flex: 1;
    display: flex;
    overflow: hidden;
  }

  .app-sidebar {
    width: 280px;
    background: var(--bg-sidebar);
    color: var(--sidebar-text-main);
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding: 24px 20px;
    overflow-y: auto;
  }

  .sidebar-block {
    background: var(--bg-sidebar-card);
    border-radius: 12px;
    padding: 16px;
    border: 1px solid var(--sidebar-border);
  }

  .sidebar-block h3 {
    margin: 0 0 12px 0;
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--sidebar-text-main);
  }

  .output-selector {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .output-path {
    background: var(--sidebar-input-bg);
    border: 1px solid var(--sidebar-input-border);
    border-radius: 8px;
    padding: 10px 12px;
    font-size: 0.85rem;
    word-break: break-all;
  }

  .sidebar-hint {
    margin: 8px 0 0;
    font-size: 0.75rem;
    color: var(--sidebar-text-sub);
  }

  .config-grid {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .config-item label {
    display: block;
    font-size: 0.8rem;
    color: var(--sidebar-text-main);
    margin-bottom: 8px;
  }

  .config-range {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .config-item input,
  .config-item select {
    width: 100%;
    padding: 8px 10px;
    border-radius: 8px;
    border: 1px solid var(--sidebar-input-border);
    background: var(--sidebar-input-bg);
    color: var(--sidebar-text-main);
    font-size: 0.85rem;
    box-sizing: border-box;
  }

  .config-item input:focus,
  .config-item select:focus {
    outline: none;
    border-color: #6366f1;
    box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.28);
  }

  .config-item select {
    appearance: none;
    -webkit-appearance: none;
    -moz-appearance: none;
    padding-right: 34px;
    background-image: linear-gradient(45deg, transparent 50%, var(--sidebar-text-sub) 50%),
      linear-gradient(135deg, var(--sidebar-text-sub) 50%, transparent 50%);
    background-position: calc(100% - 18px) calc(50% - 3px), calc(100% - 12px) calc(50% - 3px);
    background-size: 6px 6px, 6px 6px;
    background-repeat: no-repeat;
    cursor: pointer;
  }

  .config-item select option {
    background: var(--bg-panel);
    color: var(--text-main);
  }

  .helper-list {
    margin: 0;
    padding-left: 20px;
    font-size: 0.8rem;
    color: #e5e7eb;
    line-height: 1.6;
  }

  .app-content {
    flex: 1;
    overflow-y: auto;
    padding: 24px 32px;
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .card {
    background: var(--bg-panel);
    border-radius: 14px;
    border: 1px solid var(--line);
    box-shadow: 0 1px 3px rgba(15, 23, 42, 0.08);
    padding: 24px;
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    gap: 12px;
    flex-wrap: wrap;
  }

  .card-actions {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
  }

  .card-title {
    margin: 0;
    font-size: 1.2rem;
    font-weight: 600;
  }

  .card-subtitle {
    font-size: 0.85rem;
    color: var(--text-sub);
  }

  .form-grid {
    display: grid;
    grid-template-columns: 160px 1fr;
    gap: 12px 20px;
    align-items: center;
  }

  .form-label {
    font-size: 0.9rem;
    font-weight: 500;
  }

  .form-grid input[type="text"] {
    padding: 10px 14px;
    border-radius: 8px;
    border: 1px solid var(--line);
    font-size: 0.9rem;
    box-sizing: border-box;
  }

  .form-grid input[type="text"]:focus {
    outline: 2px solid #6366f1;
    border-color: transparent;
  }

  .path-row {
    display: flex;
    gap: 12px;
    align-items: center;
  }

  .path-row input {
    flex: 1;
  }

  .btn {
    border: none;
    border-radius: 8px;
    font-size: 0.9rem;
    font-weight: 600;
    padding: 10px 18px;
    cursor: pointer;
    transition: all 0.2s ease;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-primary {
    background: #4f46e5;
    color: #ffffff;
  }

  .btn-primary:hover:not(:disabled) {
    background: #4338ca;
  }

  .btn-secondary {
    background: #e5e7eb;
    color: #111827;
  }

  .btn-secondary:hover:not(:disabled) {
    background: #d1d5db;
  }

  .btn-outline {
    background: transparent;
    color: #4f46e5;
    border: 1px solid #4f46e5;
  }

  .btn-ghost {
    background: transparent;
    border: 1px solid var(--line);
    color: var(--text-main);
  }

  .btn-ghost:hover:not(:disabled) {
    background: rgba(99, 102, 241, 0.08);
  }

  .btn-outline:hover:not(:disabled) {
    background: rgba(79, 70, 229, 0.08);
  }

  .floating-hints {
    position: fixed;
    top: 86px;
    left: 50%;
    transform: translateX(-50%);
    z-index: 1200;
    display: flex;
    flex-direction: column;
    gap: 10px;
    pointer-events: none;
    width: min(560px, calc(100vw - 24px));
  }

  .hint-toast {
    border-radius: 10px;
    padding: 14px 18px;
    font-size: 0.98rem;
    line-height: 1.5;
    border: 1px solid transparent;
    box-shadow: 0 12px 30px rgba(15, 23, 42, 0.24);
    backdrop-filter: blur(3px);
    animation: toastIn 0.18s ease;
    text-align: center;
  }

  .hint-error {
    background: rgba(254, 242, 242, 0.94);
    border-color: #fecaca;
    color: #b91c1c;
  }

  .hint-success {
    background: rgba(240, 253, 244, 0.94);
    border-color: #86efac;
    color: #166534;
  }

  .theme-dark .hint-error {
    background: rgba(127, 29, 29, 0.9);
    border-color: #b91c1c;
    color: #fecaca;
  }

  .theme-dark .hint-success {
    background: rgba(20, 83, 45, 0.9);
    border-color: #16a34a;
    color: #bbf7d0;
  }

  @keyframes toastIn {
    from {
      opacity: 0;
      transform: translateY(-6px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .empty-state {
    text-align: center;
    padding: 60px 20px;
    color: #9ca3af;
  }

  .empty-icon {
    font-size: 3rem;
    margin-bottom: 12px;
  }

  .empty-hint {
    margin-top: 6px;
    font-size: 0.85rem;
  }

  .task-card {
    border: 1px solid var(--line);
    border-radius: 12px;
    padding: 20px;
    margin-bottom: 16px;
    background: var(--bg-panel);
    box-shadow: 0 1px 2px rgba(15, 23, 42, 0.05);
  }

  .task-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 20px;
  }

  .task-info {
    display: flex;
    flex-direction: column;
    gap: 6px;
    flex: 1;
  }

  .task-url {
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--text-main);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .task-actions {
    display: flex;
    gap: 8px;
  }

  .btn-icon {
    background: #e5e7eb;
    border: none;
    border-radius: 50%;
    width: 36px;
    height: 36px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
  }

  .btn-icon:hover {
    background: #d1d5db;
  }

  .btn-danger {
    background: #fee2e2;
    color: #b91c1c;
  }

  .btn-danger:hover {
    background: #fecaca;
  }

  .badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 4px 12px;
    border-radius: 999px;
    font-size: 0.75rem;
    font-weight: 600;
    width: fit-content;
  }

  .badge-pending {
    background: #fef3c7;
    color: #b45309;
  }

  .badge-running {
    background: #dbeafe;
    color: #1d4ed8;
  }

  .badge-completed {
    background: #dcfce7;
    color: #047857;
  }

  .badge-failed {
    background: #fee2e2;
    color: #b91c1c;
  }

  .badge-cancelled {
    background: #e5e7eb;
    color: #4b5563;
  }

  .task-book-title {
    margin-top: 12px;
    font-size: 0.9rem;
    font-weight: 500;
    color: var(--text-main);
  }

  .task-meta {
    margin-top: 8px;
    font-size: 0.8rem;
    color: var(--text-sub);
    display: flex;
    gap: 8px;
    align-items: baseline;
  }

  .meta-label {
    font-weight: 600;
  }

  .meta-value {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .task-progress {
    margin-top: 16px;
  }

  .progress-doc {
    font-size: 0.85rem;
    color: var(--text-sub);
    margin-bottom: 8px;
  }

  .progress-bar-container {
    position: relative;
    height: 10px;
    background: #243246;
    border-radius: 12px;
    overflow: hidden;
  }

  .progress-bar {
    position: absolute;
    top: 0;
    left: 0;
    bottom: 0;
    background: linear-gradient(90deg, #6366f1 0%, #a855f7 100%);
  }

  .progress-summary {
    margin-top: 8px;
    display: flex;
    justify-content: space-between;
    font-size: 0.8rem;
    color: var(--text-sub);
  }

  .task-error {
    margin-top: 12px;
    padding: 10px 12px;
    background: #fef2f2;
    color: #b91c1c;
    border-radius: 8px;
    font-size: 0.85rem;
  }

  .task-footer {
    margin-top: 16px;
    padding-top: 12px;
    border-top: 1px solid var(--line);
    font-size: 0.75rem;
    color: var(--text-sub);
    display: flex;
    gap: 16px;
  }

  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(17, 24, 39, 0.45);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 24px;
  }

  .modal-content {
    background: #ffffff;
    border-radius: 16px;
    width: 90%;
    max-width: 600px;
    max-height: 80vh;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .modal-header,
  .modal-footer {
    padding: 20px 24px;
    border-bottom: 1px solid #e5e7eb;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .modal-footer {
    border-bottom: none;
    border-top: 1px solid #e5e7eb;
    justify-content: flex-end;
    gap: 12px;
  }

  .modal-header h3 {
    margin: 0;
    font-size: 1.1rem;
  }

  .modal-close {
    background: none;
    border: none;
    font-size: 1.8rem;
    color: #6b7280;
    cursor: pointer;
    line-height: 1;
  }

  .modal-close:hover {
    color: #111827;
  }

  .modal-body {
    padding: 20px 24px;
    overflow-y: auto;
    flex: 1;
  }

  .modal-hint {
    font-size: 0.85rem;
    color: #6b7280;
    line-height: 1.6;
    margin-bottom: 16px;
  }

  .batch-textarea {
    width: 100%;
    padding: 12px;
    border-radius: 8px;
    border: 1px solid #d1d5db;
    font-family: 'JetBrains Mono', 'Consolas', monospace;
    font-size: 0.85rem;
    resize: vertical;
    box-sizing: border-box;
  }

  .batch-textarea:focus {
    outline: 2px solid #6366f1;
    border-color: transparent;
  }
</style>
