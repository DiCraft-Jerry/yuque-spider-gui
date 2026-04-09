<script>
  import { onMount } from 'svelte';
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

  let config = {
    delayMin: 1,
    delayMax: 4,
    timeout: 30,
    maxRetries: 3,
    concurrentDownloads: 1
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
        successMessage = '已更新默认输出目录';
        errorMessage = '';
        setTimeout(() => successMessage = '', 2500);
      }
    } catch (err) {
      console.error('选择目录失败:', err);
      errorMessage = '选择目录失败';
    }
  }

  async function addTask() {
    errorMessage = '';
    successMessage = '';

    const isValid = await ValidateURL(newTask.url);
    if (!isValid) {
      errorMessage = '请输入有效的语雀 URL';
      return;
    }

    const targetOutputPath = newTask.outputPath || defaultOutputPath;
    if (!targetOutputPath) {
      errorMessage = '请选择输出目录';
      return;
    }

    try {
      await AddTask(newTask.url, newTask.cookie, targetOutputPath, config);
      successMessage = '任务添加成功';

      defaultOutputPath = targetOutputPath;
      newTask.outputPath = targetOutputPath;
      newTask.url = '';
      newTask.cookie = '';

      setTimeout(() => successMessage = '', 3000);
    } catch (err) {
      errorMessage = '添加任务失败: ' + err;
    }
  }

  async function removeTask(taskId) {
    try {
      await RemoveTask(taskId);
    } catch (err) {
      errorMessage = '删除任务失败: ' + err;
    }
  }

  async function startTask(taskId) {
    try {
      await StartTask(taskId);
    } catch (err) {
      errorMessage = '启动任务失败: ' + err;
    }
  }

  async function cancelTask(taskId) {
    try {
      await CancelTask(taskId);
    } catch (err) {
      errorMessage = '取消任务失败: ' + err;
    }
  }

  async function startAllPending() {
    try {
      await StartAllPendingTasks();
      successMessage = '已启动所有待处理任务';
      setTimeout(() => successMessage = '', 3000);
    } catch (err) {
      errorMessage = '启动任务失败: ' + err;
    }
  }

  async function clearCompleted() {
    try {
      await ClearCompletedTasks();
      successMessage = '已清除完成或失败的任务';
      setTimeout(() => successMessage = '', 3000);
    } catch (err) {
      errorMessage = '清除任务失败: ' + err;
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
    errorMessage = '';

    const targetOutputPath = newTask.outputPath || defaultOutputPath;
    if (!targetOutputPath) {
      errorMessage = '请先选择输出目录';
      return;
    }

    const batchTasks = parseBatchInput();

    if (batchTasks.length === 0) {
      errorMessage = '没有有效的任务';
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

    successMessage = `成功添加 ${successCount}/${batchTasks.length} 个任务`;
    setTimeout(() => successMessage = '', 3000);

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
</script>

<main class="admin-app">
  <header class="app-header">
    <div class="brand">
      <div class="brand-title">语雀下载器</div>
      <div class="brand-subtitle">Yuque Spider Desktop</div>
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
      <button on:click={() => showBatchModal = true} class="btn btn-outline">批量导入</button>
      <button on:click={startAllPending} disabled={stats.pending === 0} class="btn btn-primary">开始全部</button>
      <button on:click={clearCompleted} disabled={stats.completed === 0 && stats.failed === 0} class="btn btn-secondary">清除完成</button>
    </div>
  </header>

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
        </div>
      </section>

      <section class="sidebar-block">
        <h3>快速指引</h3>
        <ul class="helper-list">
          <li>1. 选择输出目录 (仅需一次)</li>
          <li>2. 粘贴知识库 URL 与 Cookie</li>
          <li>3. 每个知识库在目录下生成独立文件夹</li>
        </ul>
      </section>
    </aside>

    <section class="app-content">
      {#if errorMessage}
        <div class="alert alert-error">❌ {errorMessage}</div>
      {/if}
      {#if successMessage}
        <div class="alert alert-success">✅ {successMessage}</div>
      {/if}

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
          <div class="card-subtitle">共 {stats.total} 个任务</div>
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
    display: flex;
    flex-direction: column;
    height: 100vh;
  }

  .app-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 28px;
    background: #ffffff;
    border-bottom: 1px solid #e5e7eb;
    box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
  }

  .brand-title {
    font-size: 1.4rem;
    font-weight: 700;
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
    color: #6b7280;
  }

  .header-actions {
    display: flex;
    gap: 12px;
  }

  .app-shell {
    flex: 1;
    display: flex;
    overflow: hidden;
  }

  .app-sidebar {
    width: 280px;
    background: #111827;
    color: #d1d5db;
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding: 24px 20px;
    overflow-y: auto;
  }

  .sidebar-block {
    background: rgba(31, 41, 55, 0.65);
    border-radius: 12px;
    padding: 16px;
  }

  .sidebar-block h3 {
    margin: 0 0 12px 0;
    font-size: 0.95rem;
    font-weight: 600;
    color: #ffffff;
  }

  .output-selector {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .output-path {
    background: rgba(55, 65, 81, 0.6);
    border-radius: 8px;
    padding: 10px 12px;
    font-size: 0.85rem;
    word-break: break-all;
  }

  .sidebar-hint {
    margin: 8px 0 0;
    font-size: 0.75rem;
    color: #9ca3af;
  }

  .config-grid {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .config-item label {
    display: block;
    font-size: 0.8rem;
    color: #d1d5db;
    margin-bottom: 8px;
  }

  .config-range {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .config-item input {
    width: 100%;
    padding: 8px 10px;
    border-radius: 8px;
    border: none;
    background: rgba(17, 24, 39, 0.6);
    color: #f3f4f6;
    font-size: 0.85rem;
  }

  .config-item input:focus {
    outline: 2px solid #6366f1;
    outline-offset: 2px;
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
    background: #ffffff;
    border-radius: 14px;
    border: 1px solid #e5e7eb;
    box-shadow: 0 1px 3px rgba(15, 23, 42, 0.08);
    padding: 24px;
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    margin-bottom: 16px;
  }

  .card-title {
    margin: 0;
    font-size: 1.2rem;
    font-weight: 600;
  }

  .card-subtitle {
    font-size: 0.85rem;
    color: #6b7280;
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
    border: 1px solid #d1d5db;
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

  .btn-outline:hover:not(:disabled) {
    background: rgba(79, 70, 229, 0.08);
  }

  .alert {
    border-radius: 10px;
    padding: 12px 16px;
    font-size: 0.9rem;
  }

  .alert-error {
    background: #fef2f2;
    border: 1px solid #fecaca;
    color: #b91c1c;
  }

  .alert-success {
    background: #f0fdf4;
    border: 1px solid #86efac;
    color: #166534;
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
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    padding: 20px;
    margin-bottom: 16px;
    background: #ffffff;
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
    color: #1f2937;
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
    color: #374151;
  }

  .task-meta {
    margin-top: 8px;
    font-size: 0.8rem;
    color: #6b7280;
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
    color: #4b5563;
    margin-bottom: 8px;
  }

  .progress-bar-container {
    position: relative;
    height: 10px;
    background: #e5e7eb;
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
    color: #6b7280;
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
    border-top: 1px solid #f3f4f6;
    font-size: 0.75rem;
    color: #9ca3af;
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
