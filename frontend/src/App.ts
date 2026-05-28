/**
 * App.ts — Static three-panel workspace skeleton (Task 7) + interactive wiring (Task 8)
 *
 * Renders the modern Wails UI shell with:
 *   - Topbar (title, version, address input, destination)
 *   - Left panel (package input card, summary card, package table)
 *   - Right panel (quote cards, recommendation card)
 *
 * Task 8 adds Wails backend bindings: GetInitialState, AddPackage, DeletePackage,
 * ClearPackages, and reactive state rendering.
 */

import { main } from './wailsjs/go/models';
import {
  GetInitialState,
  AddPackage,
  DeletePackage,
  ClearPackages,
  SetDestination,
  AnalyzeAddress,
  GetProvinces,
} from './api';

/**
 * renderSummary — pure function that serialises key state fields.
 * Exported for unit testing (check-summary-test.ts).
 */
export function renderSummary(state: main.QuoteState): string {
  return `${state.summary.totalCount}|${state.sto.price}|${state.bs.price}`;
}

/**
 * shouldAnalyzeAddress — returns true when the trimmed address length >= 5.
 */
export function shouldAnalyzeAddress(value: string): boolean {
  return value.trim().length >= 5;
}

/**
 * buildStoNote — returns an overweight warning string when applicable.
 */
export function buildStoNote(weight: number, available: boolean): string {
  if (!available && weight > 50) return '⚠ >50kg 仅发百世';
  return '';
}

/**
 * setStatus — shows or hides the status-message bar.
 * When message is empty/falsy the bar is hidden; otherwise shown with given text.
 */
function setStatus(root: HTMLElement, message: string): void {
  const el = root.querySelector<HTMLElement>('.status-message');
  if (!el) return;
  if (message) {
    el.textContent = message;
    el.style.display = 'block';
  } else {
    el.style.display = 'none';
  }
}

/**
 * renderFromState — updates the DOM with state from the Wails backend.
 * Targets specific sections: summary, package table, quote cards, recommendation.
 * Null-safe: guards against missing summary/sto/bs on the state object.
 */
function renderFromState(root: HTMLElement, state: main.QuoteState): void {
  // Clear any previous status messages on successful render
  setStatus(root, '');

  const summary = state?.summary;
  const sto = state?.sto;
  const bs = state?.bs;
  const dest = state?.destination;

  // --- Destination dropdowns ---
  if (dest) {
    // Province select
    const provSelect = root.querySelector<HTMLSelectElement>('.province-select');
    if (provSelect && dest.province) {
      provSelect.value = dest.province;
    }

    // City select — rebuild options
    const citySelect = root.querySelector<HTMLSelectElement>('.city-select');
    if (citySelect) {
      citySelect.innerHTML = '<option value="">请选择城市</option>';
      if (dest.cities && dest.cities.length > 0) {
        dest.cities.forEach(c => {
          const opt = document.createElement('option');
          opt.value = c;
          opt.textContent = c;
          if (c === dest.city) opt.selected = true;
          citySelect.appendChild(opt);
        });
      }
      if (dest.city) {
        citySelect.value = dest.city;
      }
    }
  }

  // --- Summary ---
  const countEl = root.querySelector('.summary-count');
  if (countEl && summary) countEl.textContent = String(summary.totalCount ?? 0);
  const weightEl = root.querySelector('.summary-weight');
  if (weightEl && summary) weightEl.textContent = (summary.displayWeight ?? 0).toFixed(2);

  // --- Quote cards ---
  const quoteCards = root.querySelectorAll('.quote-card');
  if (quoteCards.length >= 1 && sto) {
    const stoPrice = quoteCards[0].querySelector('.quote-price');
    const stoLabel = quoteCards[0].querySelector('.detail-label');
    const stoNote = quoteCards[0].querySelector('.detail-note');
    if (stoPrice) stoPrice.textContent = `¥${(sto.price ?? 0).toFixed(2)}`;
    if (stoLabel) stoLabel.textContent = `计费重量：${(sto.billableWeight ?? 0).toFixed(2)} kg`;
    // Apply buildStoNote to show overweight hints
    const stoOverweightNote = buildStoNote(sto.billableWeight ?? 0, sto.available);
    if (stoNote) stoNote.textContent = stoOverweightNote || (sto.available ? (sto.note || '可报价') : (sto.note || '暂未报价'));
  }
  if (quoteCards.length >= 2 && bs) {
    const bsPrice = quoteCards[1].querySelector('.quote-price');
    const bsLabel = quoteCards[1].querySelector('.detail-label');
    const bsNote = quoteCards[1].querySelector('.detail-note');
    if (bsPrice) bsPrice.textContent = `¥${(bs.price ?? 0).toFixed(2)}`;
    if (bsLabel) bsLabel.textContent = `计费重量：${(bs.billableWeight ?? 0).toFixed(2)} kg`;
    if (bsNote) bsNote.textContent = bs.available ? (bs.note || '可报价') : (bs.note || '暂未报价');
  }

  // --- Recommendation ---
  const recText = root.querySelector('.rec-text');
  const recBadge = root.querySelector('.rec-badge');
  if (recText) {
    if (sto?.recommended) {
      recText.innerHTML = '✅ <strong>申通快递</strong> — 价格最优';
      if (recBadge) recBadge.textContent = '申通快递';
    } else if (bs?.recommended) {
      recText.innerHTML = '✅ <strong>百世快运</strong> — 价格最优';
      if (recBadge) recBadge.textContent = '百世快运';
    } else {
      recText.textContent = '请先添加包裹并设置目的地';
      if (recBadge) recBadge.textContent = '';
    }
  }

  // --- Package table ---
  const tbody = root.querySelector('.package-table tbody');
  if (tbody) {
    if (!state.packages || state.packages.length === 0) {
      tbody.innerHTML = '<tr><td colspan="7" class="empty-row">暂无包裹，请点击"添加"录入</td></tr>';
    } else {
      tbody.innerHTML = state.packages.map(p => {
        const dims = (p.length && p.width && p.height)
          ? `${p.length}×${p.width}×${p.height}`
          : '-';
        const billable = Math.max(p.stoBillable, p.bsBillable);
        return `<tr>
          <td>${p.id}</td>
          <td>${p.modeLabel}</td>
          <td>${dims}</td>
          <td>${p.actualWeight.toFixed(2)}</td>
          <td>${billable.toFixed(2)}</td>
          <td>${p.quantity}</td>
          <td><button class="btn-delete" data-id="${p.id}">删除</button></td>
        </tr>`;
      }).join('');
    }
  }
}

export function renderShell(): string {
  return `
<div class="app-shell">
<header class="topbar">
  <div class="topbar-inner">
    <div class="topbar-left">
      <h1 class="app-title">快递运费报价</h1>
      <span class="app-version">v3.0</span>
    </div>
    <div class="topbar-center">
      <div class="address-group">
        <input type="text" class="address-input" placeholder="请输入收件地址（自动解析），如：北京市朝阳区..." />
      </div>
    </div>
    <div class="topbar-right">
      <span class="destination-label">目的地：</span>
      <select class="province-select">
        <option value="">请选择省份</option>
      </select>
      <select class="city-select">
        <option value="">请选择城市</option>
      </select>
    </div>
  </div>
</header>
<div class="status-message" style="display:none"></div>
<main class="workspace">
  <section class="packages-panel">
    <div class="card package-input-card">
      <h2 class="card-title">包裹录入</h2>
      <div class="input-row">
        <div class="field">
          <label>长 (cm)</label>
          <input type="number" class="input-dim" placeholder="长" />
        </div>
        <div class="field">
          <label>宽 (cm)</label>
          <input type="number" class="input-dim" placeholder="宽" />
        </div>
        <div class="field">
          <label>高 (cm)</label>
          <input type="number" class="input-dim" placeholder="高" />
        </div>
      </div>
      <div class="dim-weight-hint">默认体积计算方式：长×宽×高(cm³)，计费重=体积÷系数(申通8000/百世5000)</div>
      <div class="input-row">
        <div class="field">
          <label>实际重量 (kg)</label>
          <input type="number" step="0.1" class="input-weight" placeholder="重量" />
        </div>
        <div class="field">
          <label>数量</label>
          <input type="number" class="input-qty" placeholder="1" value="1" />
        </div>
        <div class="field field-btn">
          <button class="btn btn-add">添加</button>
        </div>
      </div>
    </div>

    <div class="card package-preview-card">
      <div class="preview-summary">
        <span class="preview-item">当前包裹：<strong class="summary-count">0</strong> 个</span>
        <span class="preview-divider">|</span>
        <span class="preview-item">总重量：<strong class="summary-weight">0.00</strong> kg</span>
      </div>
      <button class="btn btn-clear">清空全部</button>
    </div>

    <div class="card package-list-card">
      <table class="package-table">
        <thead>
          <tr>
            <th>#</th>
            <th>模式</th>
            <th>尺寸 (cm)</th>
            <th>实际重量 (kg)</th>
            <th>计费重量 (kg)</th>
            <th>数量</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td colspan="7" class="empty-row">暂无包裹，请点击"添加"录入</td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>

  <aside class="quotes-panel">
    <div class="card quote-card">
      <div class="quote-header">
        <h3 class="quote-name">申通快递</h3>
        <span class="quote-provider">STO Express</span>
      </div>
      <div class="quote-price">¥0.00</div>
      <div class="quote-detail">
        <span class="detail-label">计费重量：0.00 kg</span>
        <span class="detail-note">暂未报价</span>
      </div>
    </div>

    <div class="card quote-card">
      <div class="quote-header">
        <h3 class="quote-name">百世快运</h3>
        <span class="quote-provider">Best Express</span>
      </div>
      <div class="quote-price">¥0.00</div>
      <div class="quote-detail">
        <span class="detail-label">计费重量：0.00 kg</span>
        <span class="detail-note">暂未报价</span>
      </div>
    </div>

    <div class="card recommendation-card">
      <h3 class="rec-title">推荐选择</h3>
      <span class="rec-badge"></span>
      <p class="rec-text">请先添加包裹并设置目的地</p>
    </div>
  </aside>
</main>
</div>`;
}

export function mountApp(root: HTMLElement): void {
  root.innerHTML = renderShell();

  async function refresh(): Promise<void> {
    try {
      setStatus(root, '');
      const state = await GetInitialState();
      renderFromState(root, state);
    } catch (err) {
      setStatus(root, '⚠️ 连接后端失败，请检查服务状态');
      console.error('Failed to get state:', err);
    }
  }

  // Load initial state from backend
  refresh();

  // --- Populate province select ---
  (async () => {
    try {
      const provinces = await GetProvinces();
      const provSelect = root.querySelector<HTMLSelectElement>('.province-select');
      if (provSelect) {
        provinces.forEach(p => {
          const opt = document.createElement('option');
          opt.value = p;
          opt.textContent = p;
          provSelect.appendChild(opt);
        });
      }
    } catch (err) {
      console.error('Failed to load provinces:', err);
    }
  })();

  // --- 300ms debounced address input with auto-parse ---
  // Note: addressTimer lives for the lifetime of the app (no cleanup needed).
  let addressTimer: ReturnType<typeof setTimeout> | null = null;
  const addressInput = root.querySelector<HTMLInputElement>('.address-input');

  // Request counter to discard stale responses from rapid clicks/auto-trigger.
  let analyzeRequestId = 0;

  // Shared analyze-and-apply logic
  const analyzeAndApply = async (rawValue: string) => {
    const value = rawValue.trim();
    if (!shouldAnalyzeAddress(value)) {
      setStatus(root, '');
      return;
    }
    const thisRequestId = ++analyzeRequestId;
    try {
      setStatus(root, '🔍 正在解析地址...');
      const result = await AnalyzeAddress(value);
      if (thisRequestId !== analyzeRequestId) return;
      if (!result.province && !result.city) {
        setStatus(root, '⚠️ 地址解析失败，请补充更详细的信息');
        return;
      }
      setStatus(root, '');
      const state = await SetDestination(result.province, result.city);
      if (thisRequestId !== analyzeRequestId) return;
      renderFromState(root, state);
    } catch (err: any) {
      if (thisRequestId !== analyzeRequestId) return;
      const msg = String(err?.message || err || '');
      if (msg.includes('DEEPSEEK_API_KEY') || msg.includes('API key') || msg.includes('api_key')) {
        setStatus(root, '⚠️ DEEPSEEK_API_KEY 未配置，AI 解析不可用');
      } else {
        setStatus(root, '⚠️ 地址解析失败，请重试或手动选择');
      }
      console.error('AnalyzeAddress failed:', err);
    }
  };

  const updateDebouncedAddress = () => {
    if (addressTimer) clearTimeout(addressTimer);
    addressTimer = setTimeout(() => {
      const rawValue = addressInput?.value ?? '';
      analyzeAndApply(rawValue);
    }, 300);
  };
  addressInput?.addEventListener('input', updateDebouncedAddress);
  addressInput?.addEventListener('change', updateDebouncedAddress);

  // --- Province select change ---
  root.querySelector('.province-select')?.addEventListener('change', async (e) => {
    const province = (e.target as HTMLSelectElement).value;
    try {
      setStatus(root, '');
      const state = await SetDestination(province, '');
      renderFromState(root, state);
    } catch (err) {
      setStatus(root, '⚠️ 设置目的地失败');
      console.error('SetDestination failed:', err);
    }
  });

  // --- City select change ---
  root.querySelector('.city-select')?.addEventListener('change', async (e) => {
    const provSelect = root.querySelector<HTMLSelectElement>('.province-select');
    const province = provSelect?.value || '';
    const city = (e.target as HTMLSelectElement).value;
    try {
      setStatus(root, '');
      const state = await SetDestination(province, city);
      renderFromState(root, state);
    } catch (err) {
      setStatus(root, '⚠️ 设置目的地失败');
      console.error('SetDestination failed:', err);
    }
  });

  // --- Add package (shared logic) ---
  const addPackageFromInputs = async () => {
    const dims = root.querySelectorAll<HTMLInputElement>('.input-dim');
    const length = parseFloat(dims[0]?.value || '0');
    const width = parseFloat(dims[1]?.value || '0');
    const height = parseFloat(dims[2]?.value || '0');
    const actualWeight = parseFloat(
      (root.querySelector<HTMLInputElement>('.input-weight') as HTMLInputElement | null)?.value || '0'
    );
    const quantity = parseInt(
      (root.querySelector<HTMLInputElement>('.input-qty') as HTMLInputElement | null)?.value || '1', 10
    );

    // Validate: reject negatives and zero quantity
    if (length < 0 || width < 0 || height < 0) {
      setStatus(root, '⚠️ 尺寸不能为负数');
      return;
    }
    if (actualWeight < 0) {
      setStatus(root, '⚠️ 重量不能为负数');
      return;
    }
    if (quantity < 1) {
      setStatus(root, '⚠️ 数量至少为 1');
      return;
    }

    const input = new main.PackageInput({
      length,
      width,
      height,
      volume: 0,
      actualWeight,
      quantity: quantity > 0 ? quantity : 1,
    });

    try {
      setStatus(root, '');
      const state = await AddPackage(input);
      renderFromState(root, state);
    } catch (err) {
      setStatus(root, '⚠️ 连接后端失败，请检查服务状态');
      console.error('Failed to add package:', err);
    }
  };

  root.querySelector('.btn-add')?.addEventListener('click', addPackageFromInputs);

  // --- Enter key on package input fields triggers add ---
  root.querySelectorAll('.input-dim, .input-weight, .input-qty').forEach(el => {
    el.addEventListener('keydown', (e) => {
      if ((e as KeyboardEvent).key === 'Enter') addPackageFromInputs();
    });
  });

  // --- Clear all packages ---
  root.querySelector('.btn-clear')?.addEventListener('click', async () => {
    try {
      setStatus(root, '');
      const state = await ClearPackages();
      renderFromState(root, state);
    } catch (err) {
      setStatus(root, '⚠️ 连接后端失败，请检查服务状态');
      console.error('Failed to clear packages:', err);
    }
  });

  // --- Delete single package (delegated on table) ---
  root.querySelector('.package-table')?.addEventListener('click', async (e) => {
    const btn = (e.target as HTMLElement).closest('.btn-delete');
    if (!btn) return;
    const id = parseInt(btn.getAttribute('data-id') || '0', 10);
    if (id <= 0) return;
    try {
      setStatus(root, '');
      const state = await DeletePackage(id);
      renderFromState(root, state);
    } catch (err) {
      setStatus(root, '⚠️ 连接后端失败，请检查服务状态');
      console.error('Failed to delete package:', err);
    }
  });

  // --- Click-to-copy on .quote-price ---
  root.querySelector('.quotes-panel')?.addEventListener('click', (e) => {
    const priceEl = (e.target as HTMLElement).closest('.quote-price');
    if (!priceEl) return;
    const priceText = priceEl.textContent || '';
    const numericPart = priceText.replace('¥', '').trim();
    if (!numericPart) return;

    const doCopy = (text: string): Promise<void> => {
      // Prefer modern Clipboard API
      if (typeof navigator.clipboard?.writeText === 'function') {
        return navigator.clipboard.writeText(text);
      }
      // Fallback: execCommand('copy') via temp input
      return new Promise<void>((resolve, reject) => {
        try {
          const input = document.createElement('input');
          input.value = text;
          input.style.position = 'fixed';
          input.style.opacity = '0';
          document.body.appendChild(input);
          input.select();
          const ok = document.execCommand('copy');
          document.body.removeChild(input);
          if (ok) resolve(); else reject(new Error('execCommand copy failed'));
        } catch (err) {
          reject(err);
        }
      });
    };

    doCopy(numericPart).then(() => {
      const brief = `已复制 ¥${numericPart}`;
      setStatus(root, brief);
      setTimeout(() => {
        const currentMsg = root.querySelector('.status-message')?.textContent || '';
        if (currentMsg === brief) setStatus(root, '');
      }, 1500);
    }).catch(() => {
      setStatus(root, '⚠️ 复制功能不可用');
    });
  });
}
