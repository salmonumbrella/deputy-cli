package auth

const setupTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Deputy CLI - Connect Your Account</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Poppins:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg: #F7F8FC;
            --bg-card: #FFFFFF;
            --bg-hint: #F0F1F8;
            --bg-input: #F5F6FB;
            --border: #E2E4EF;
            --border-focus: #4429BA;
            --text: #1A1B2E;
            --text-secondary: #5C5E7A;
            --text-muted: #9496AD;
            --primary: #0C017B;
            --primary-hover: #4429BA;
            --primary-light: #ECEAFD;
            --accent: #7F52FD;
            --accent-light: #F3EFFF;
            --teal: #37CFCD;
            --teal-light: #E6FAF9;
            --success: #10B981;
            --success-light: #D1FAE5;
            --error: #EF4444;
            --error-light: #FEE2E2;
        }

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        html { height: 100%; }

        body {
            font-family: 'Poppins', -apple-system, BlinkMacSystemFont, sans-serif;
            background: var(--bg);
            color: var(--text);
            min-height: 100%;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 2rem 1.5rem 3rem;
            position: relative;
            overflow-x: hidden;
        }

        /* Decorative background pattern */
        body::before {
            content: '';
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background:
                linear-gradient(135deg, rgba(12, 1, 123, 0.03) 0%, transparent 50%),
                linear-gradient(225deg, rgba(55, 207, 205, 0.04) 0%, transparent 50%);
            pointer-events: none;
            z-index: 0;
        }

        /* Geometric accent shapes */
        body::after {
            content: '';
            position: fixed;
            top: -20%;
            right: -10%;
            width: 500px;
            height: 500px;
            background: radial-gradient(circle, rgba(127, 82, 253, 0.06) 0%, transparent 70%);
            pointer-events: none;
            z-index: 0;
        }

        .container {
            width: 100%;
            max-width: 400px;
            position: relative;
            z-index: 1;
        }

        /* Logo */
        .logo {
            display: flex;
            justify-content: center;
            margin-bottom: 1rem;
            animation: fadeDown 0.5s ease-out;
        }

        .logo svg {
            height: 28px;
            width: auto;
        }

        @keyframes fadeDown {
            from { opacity: 0; transform: translateY(-10px); }
            to { opacity: 1; transform: translateY(0); }
        }

        .logo-text {
            font-size: 1.75rem;
            font-weight: 700;
            color: var(--primary);
            letter-spacing: -0.02em;
        }

        .logo-text span {
            color: var(--teal);
        }

        /* CLI Badge */
        .badge-wrapper {
            display: flex;
            justify-content: center;
            margin-bottom: 1.5rem;
            animation: fadeDown 0.5s ease-out 0.1s both;
        }

        .cli-badge {
            display: inline-flex;
            align-items: center;
            gap: 0.375rem;
            background: var(--primary-light);
            color: var(--primary);
            font-size: 0.6875rem;
            font-weight: 600;
            padding: 0.375rem 0.75rem;
            border-radius: 100px;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }

        .cli-badge svg {
            width: 12px;
            height: 12px;
        }

        h1 {
            font-size: 1.375rem;
            font-weight: 700;
            letter-spacing: -0.02em;
            margin-bottom: 0.25rem;
            text-align: center;
            animation: fadeDown 0.5s ease-out 0.15s both;
        }

        .subtitle {
            color: var(--text-secondary);
            font-size: 0.875rem;
            margin-bottom: 1.25rem;
            text-align: center;
            animation: fadeDown 0.5s ease-out 0.2s both;
        }

        /* Credentials hint */
        .credentials-hint {
            background: var(--bg-hint);
            border-radius: 12px;
            padding: 0.875rem;
            margin-bottom: 1rem;
            animation: fadeDown 0.5s ease-out 0.25s both;
        }

        .hint-header {
            display: flex;
            align-items: center;
            gap: 0.5rem;
            font-size: 0.75rem;
            font-weight: 500;
            color: var(--text-secondary);
            margin-bottom: 0.625rem;
        }

        .hint-header svg {
            width: 14px;
            height: 14px;
            color: var(--text-muted);
        }

        .hint-link {
            display: flex;
            align-items: center;
            gap: 0.5rem;
            padding: 0.625rem 0.75rem;
            background: var(--bg-card);
            border: 1px solid var(--border);
            border-radius: 10px;
            text-decoration: none;
            color: var(--text);
            transition: all 0.2s ease;
        }

        .hint-link:hover {
            border-color: var(--accent);
            box-shadow: 0 0 0 3px rgba(127, 82, 253, 0.08);
            transform: translateY(-1px);
        }

        .hint-link-icon {
            width: 32px;
            height: 32px;
            border-radius: 8px;
            display: flex;
            align-items: center;
            justify-content: center;
            background: linear-gradient(135deg, var(--primary) 0%, var(--accent) 100%);
            color: white;
            flex-shrink: 0;
        }

        .hint-link-icon svg {
            width: 16px;
            height: 16px;
        }

        .hint-link-text { flex: 1; }

        .hint-link-title {
            font-weight: 600;
            font-size: 0.8125rem;
        }

        .hint-link-path {
            font-size: 0.6875rem;
            color: var(--text-muted);
        }

        .hint-link-arrow {
            color: var(--text-muted);
            transition: transform 0.2s ease;
        }

        .hint-link:hover .hint-link-arrow {
            transform: translateX(2px);
        }

        /* Form card */
        .form-card {
            background: var(--bg-card);
            border: 1px solid var(--border);
            border-radius: 16px;
            padding: 1.25rem;
            box-shadow: 0 4px 24px rgba(12, 1, 123, 0.04);
            animation: fadeUp 0.5s ease-out 0.3s both;
        }

        @keyframes fadeUp {
            from { opacity: 0; transform: translateY(10px); }
            to { opacity: 1; transform: translateY(0); }
        }

        .form-group {
            margin-bottom: 1rem;
        }

        .form-group:last-of-type {
            margin-bottom: 0;
        }

        .label-row {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 0.375rem;
        }

        label {
            font-size: 0.8125rem;
            font-weight: 600;
            color: var(--text);
        }

        .badge {
            font-size: 0.5625rem;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.04em;
            padding: 0.1875rem 0.5rem;
            border-radius: 4px;
            background: var(--teal-light);
            color: #0D9488;
        }

        .input-wrapper {
            position: relative;
        }

        input, select {
            width: 100%;
            padding: 0.625rem 0.875rem;
            font-family: inherit;
            font-size: 0.8125rem;
            background: var(--bg-input);
            border: 1.5px solid transparent;
            border-radius: 10px;
            color: var(--text);
            transition: all 0.2s ease;
            -webkit-appearance: none;
            appearance: none;
        }

        input::placeholder {
            color: var(--text-muted);
        }

        input:focus, select:focus {
            outline: none;
            background: var(--bg-card);
            border-color: var(--accent);
            box-shadow: 0 0 0 3px rgba(127, 82, 253, 0.12);
        }

        input.mono {
            font-family: 'JetBrains Mono', monospace;
            font-size: 0.75rem;
            letter-spacing: -0.01em;
        }

        input.error, select.error {
            border-color: var(--error);
            background: var(--error-light);
        }

        input.error:focus, select.error:focus {
            box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.12);
        }

        .input-hint {
            font-size: 0.6875rem;
            color: var(--text-muted);
            margin-top: 0.25rem;
        }

        /* Select dropdown */
        .select-wrapper {
            position: relative;
        }

        .select-wrapper::after {
            content: '';
            position: absolute;
            right: 0.875rem;
            top: 50%;
            transform: translateY(-50%);
            width: 0;
            height: 0;
            border-left: 5px solid transparent;
            border-right: 5px solid transparent;
            border-top: 5px solid var(--text-muted);
            pointer-events: none;
        }

        select {
            padding-right: 2.5rem;
            cursor: pointer;
        }

        /* Password toggle */
        .password-toggle {
            position: absolute;
            right: 0.625rem;
            top: 50%;
            transform: translateY(-50%);
            background: none;
            border: none;
            color: var(--text-muted);
            cursor: pointer;
            padding: 0.25rem;
            border-radius: 6px;
            display: flex;
            align-items: center;
            justify-content: center;
            transition: color 0.2s ease;
        }

        .password-toggle:hover {
            color: var(--accent);
        }

        .password-toggle svg {
            width: 18px;
            height: 18px;
        }

        /* Buttons */
        .btn-group {
            display: flex;
            gap: 0.625rem;
            margin-top: 1.25rem;
        }

        button {
            flex: 1;
            padding: 0.75rem 1rem;
            font-family: inherit;
            font-size: 0.8125rem;
            font-weight: 600;
            border-radius: 10px;
            cursor: pointer;
            transition: all 0.2s ease;
            border: none;
        }

        .btn-secondary {
            background: var(--bg-input);
            color: var(--text-secondary);
            border: 1px solid var(--border);
        }

        .btn-secondary:hover {
            background: var(--border);
            color: var(--text);
        }

        .btn-primary {
            background: linear-gradient(135deg, var(--primary) 0%, var(--primary-hover) 100%);
            color: white;
            box-shadow: 0 4px 12px rgba(12, 1, 123, 0.25);
        }

        .btn-primary:hover {
            transform: translateY(-1px);
            box-shadow: 0 6px 16px rgba(12, 1, 123, 0.3);
        }

        .btn-primary:active {
            transform: translateY(0);
        }

        button:disabled {
            opacity: 0.5;
            cursor: not-allowed;
            transform: none !important;
        }

        /* Status toast */
        .status {
            position: fixed;
            bottom: 2rem;
            left: 50%;
            transform: translateX(-50%) translateY(20px);
            padding: 0.75rem 1.25rem;
            border-radius: 12px;
            font-size: 0.8125rem;
            font-weight: 500;
            align-items: center;
            gap: 0.5rem;
            opacity: 0;
            visibility: hidden;
            transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
            display: flex;
            box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
            z-index: 100;
            white-space: nowrap;
        }

        .status.show {
            opacity: 1;
            visibility: visible;
            transform: translateX(-50%) translateY(0);
        }

        .status.loading {
            background: var(--primary-light);
            color: var(--primary);
        }
        .status.success {
            background: var(--success-light);
            color: var(--success);
        }
        .status.error {
            background: var(--error-light);
            color: var(--error);
        }

        .spinner {
            width: 14px;
            height: 14px;
            border: 2px solid currentColor;
            border-top-color: transparent;
            border-radius: 50%;
            animation: spin 0.7s linear infinite;
        }

        @keyframes spin { to { transform: rotate(360deg); } }

        .status-icon { width: 14px; height: 14px; flex-shrink: 0; }

        .github-link {
            margin-top: 1.25rem;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            gap: 0.5rem;
            text-decoration: none;
            color: var(--text-muted);
            font-size: 0.75rem;
            font-weight: 500;
            transition: color 0.2s ease;
            width: 100%;
        }

        .github-link:hover {
            color: var(--text-secondary);
        }

        .github-link svg {
            width: 16px;
            height: 16px;
        }

        /* Dynamic API URL hint */
        .api-url-hint {
            background: linear-gradient(135deg, var(--teal-light) 0%, var(--accent-light) 100%);
            border: 1px solid rgba(55, 207, 205, 0.2);
            border-radius: 10px;
            padding: 0.75rem;
            margin-bottom: 1rem;
            transition: all 0.3s ease;
        }

        .api-url-hint.ready {
            border-color: var(--teal);
            box-shadow: 0 0 0 3px rgba(55, 207, 205, 0.1);
        }

        .api-url-header {
            display: flex;
            align-items: center;
            gap: 0.375rem;
            font-size: 0.6875rem;
            font-weight: 600;
            color: var(--text-secondary);
            margin-bottom: 0.5rem;
            text-transform: uppercase;
            letter-spacing: 0.03em;
        }

        .api-url-header svg {
            width: 12px;
            height: 12px;
            color: var(--teal);
        }

        .api-url-link {
            display: flex;
            align-items: center;
            gap: 0.5rem;
            padding: 0.5rem 0.625rem;
            background: var(--bg-card);
            border-radius: 8px;
            text-decoration: none;
            font-family: 'JetBrains Mono', monospace;
            font-size: 0.6875rem;
            color: var(--text-muted);
            transition: all 0.2s ease;
            overflow: hidden;
        }

        .api-url-link.active {
            color: var(--primary);
            cursor: pointer;
        }

        .api-url-link.active:hover {
            background: var(--primary-light);
            transform: translateY(-1px);
        }

        .api-url-link svg {
            flex-shrink: 0;
            width: 14px;
            height: 14px;
            color: var(--teal);
        }

        .api-url-text {
            flex: 1;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .api-url-text .placeholder {
            color: var(--text-muted);
            font-style: italic;
        }

        .api-url-text .dynamic {
            color: var(--accent);
            font-weight: 500;
        }

        .api-url-arrow {
            flex-shrink: 0;
            width: 14px;
            height: 14px;
            color: var(--text-muted);
            opacity: 0;
            transform: translateX(-4px);
            transition: all 0.2s ease;
        }

        .api-url-link.active .api-url-arrow {
            opacity: 1;
            transform: translateX(0);
        }

        .api-url-link.active:hover .api-url-arrow {
            transform: translateX(2px);
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">
            <svg viewBox="0 0 182.7 43" xmlns="http://www.w3.org/2000/svg">
                <g id="full-rgb">
                    <g id="type" fill="#0C017B">
                        <path d="M69.5,8.8c-3.3,0.1-6.2,2-7.6,4.9V9.3h-5.9V43h6.1V31.8c0,0,1.5,4.6,8.4,4.6c6,0,11.1-5.5,11.1-13.7C81.6,14.6,76.5,8.8,69.5,8.8 M68.6,30.4c-3.7,0-6.5-2.8-6.5-7.6c0-4.9,2.8-7.9,6.5-7.9c3.8,0,6.4,3.2,6.4,7.9C75,27.5,72.4,30.4,68.6,30.4"/>
                        <path d="M19.4,0v12.6c-2-3.1-4.5-3.7-7.2-3.7C5.1,8.9,0,14.4,0,22.7s5.1,13.8,12.2,13.8c2.7,0,6.5-1.8,7.6-4.9v4.7h5.9V0H19.4z M13.1,30.4c-3.8,0-6.5-2.9-6.5-7.7s2.6-7.8,6.5-7.8c3.7,0,6.5,2.8,6.5,7.7C19.7,27.6,16.9,30.4,13.1,30.4"/>
                        <path d="M100.4,26.2c0,2.6-2,4.2-5.1,4.2c-2.4,0-5-1.1-5-5.2V9.3h-5.9v17.1c0,5.5,2.9,10,9.2,10c3.2,0,6.7-2.5,7.2-4.6v4.4h5.6V9.3h-6.1V26.2z"/>
                        <path d="M114.7,3.9c0,0-0.3,5.4-5.1,5.4V15h2.9v13.7c0,5.4,2.5,7.8,7.3,7.8c1.6,0,3.2-0.2,4.8-0.7v-6.2c-1.1,0.5-2.4,0.8-3.6,0.8c-1.3,0-2.4-0.6-2.4-2.3V15h9.4l6.2,20.4l-2.9,7.6h6.3l12.2-33.7h-6.5l-5.7,17.5L132,9.3h-13.4V3.9H114.7z"/>
                        <path d="M35.1,19.8h11.7c-0.2-3.4-2-5.4-5.4-5.4C37.9,14.4,35.5,16.6,35.1,19.8 M42.2,36.5c-7.6,0-13.7-5.2-13.7-13.6c0-9,5.8-14.1,12.9-14.1c6.3,0,11.6,3.9,11.6,13.7c0,0.7-0.1,1.4-0.1,2.1H35c0.3,3.9,3.8,5.8,8.2,5.8c5.7,0,8.6-2,8.6-2l0,5.8C51.9,34.2,48.6,36.5,42.2,36.5"/>
                    </g>
                    <g id="logoMark" fill="#F26A60">
                        <path d="M169,17.4C169,17.4,169,17.4,169,17.4c-0.3,0.5-0.4,0.9-0.6,1.3l-3.6,12.1h5.2l3.1-10.7c0.1-0.5,0.6-0.7,1.1-0.6c0.2,0,0.3,0.1,0.4,0.3l5.2,5.5l2.7-4.6l-4-4.2c-2.1-2.5-5.8-2.8-8.3-0.7C169.8,16.3,169.4,16.8,169,17.4"/>
                        <path d="M165.6,15.4C165.5,15.4,165.5,15.4,165.6,15.4c-0.3-0.5-0.5-0.8-0.9-1.2L156,5.1l-2.7,4.4l7.8,8.2c0.3,0.3,0.3,0.9,0,1.2c-0.1,0.1-0.3,0.2-0.4,0.2l-7.4,1.7l2.7,4.6l5.6-1.4c3.2-0.6,5.3-3.6,4.7-6.8C166.1,16.6,165.9,16,165.6,15.4"/>
                        <path d="M169,13.5L169,13.5c0.5,0,1-0.1,1.4-0.2l12.3-3l-2.4-4.6l-11,2.7c-0.5,0.1-0.9-0.2-1-0.6c0-0.2,0-0.3,0-0.5l2.2-7.2l-5.3,0l-1.6,5.5c-1.1,3.1,0.5,6.4,3.6,7.5C167.7,13.3,168.3,13.4,169,13.5"/>
                    </g>
                </g>
            </svg>
        </div>

        <div class="badge-wrapper">
            <div class="cli-badge">
                <svg viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <rect x="2" y="2" width="12" height="12" rx="2" stroke="currentColor" stroke-width="1.5"/>
                    <path d="M5 6L7 8L5 10" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                    <path d="M9 10H11" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
                </svg>
                CLI Authentication
            </div>
        </div>

        <h1>Connect Your Account</h1>
        <p class="subtitle">Link your Deputy workspace to start using the CLI</p>

        <div class="form-card">
            <form id="setupForm" autocomplete="off">
                <div class="form-group">
                    <div class="label-row">
                        <label for="install">Install Name</label>
                        <span class="badge">Required</span>
                    </div>
                    <input type="text" id="install" name="install" placeholder="e.g., mycompany" required autofocus>
                    <div class="input-hint">Your Deputy subdomain (the part before .deputy.com)</div>
                </div>

                <div class="form-group">
                    <div class="label-row">
                        <label for="geo">Geographic Region</label>
                        <span class="badge">Required</span>
                    </div>
                    <div class="select-wrapper">
                        <select id="geo" name="geo" required>
                            <option value="">Select your region...</option>
                            <option value="au">Australia (.au.deputy.com)</option>
                            <option value="uk">United Kingdom (.uk.deputy.com)</option>
                            <option value="na">North America (.na.deputy.com)</option>
                        </select>
                    </div>
                </div>

                <div class="api-url-hint" id="apiUrlHint">
                    <div class="api-url-header">
                        <svg viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
                            <path d="M13.5 6.5L8 2L2.5 6.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                            <path d="M4 5.5V13H12V5.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                            <rect x="6.5" y="9" width="3" height="4" stroke="currentColor" stroke-width="1.5"/>
                        </svg>
                        Get your API token here
                    </div>
                    <a href="#" id="apiUrlLink" class="api-url-link" target="_blank" onclick="return false;">
                        <svg viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
                            <path d="M6.5 9.5L14 2M14 2H9M14 2V7" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                            <path d="M13 9V13C13 13.5523 12.5523 14 12 14H3C2.44772 14 2 13.5523 2 13V4C2 3.44772 2.44772 3 3 3H7" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
                        </svg>
                        <span class="api-url-text" id="apiUrlText">
                            <span class="placeholder">Enter install name & region above</span>
                        </span>
                        <svg class="api-url-arrow" viewBox="0 0 16 16" fill="none">
                            <path d="M6 4L10 8L6 12" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                        </svg>
                    </a>
                </div>

                <div class="form-group">
                    <div class="label-row">
                        <label for="token">API Token</label>
                        <span class="badge">Required</span>
                    </div>
                    <div class="input-wrapper">
                        <input type="password" id="token" name="token" class="mono" placeholder="Paste your permanent API token" required style="padding-right: 2.5rem;">
                        <button type="button" class="password-toggle" id="togglePassword" aria-label="Toggle token visibility">
                            <svg id="eyeIcon" viewBox="0 0 18 18" fill="none">
                                <path d="M2 9C2 9 4.5 4 9 4C13.5 4 16 9 16 9C16 9 13.5 14 9 14C4.5 14 2 9 2 9Z" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
                                <circle cx="9" cy="9" r="2" stroke="currentColor" stroke-width="1.5"/>
                            </svg>
                            <svg id="eyeOffIcon" style="display:none" viewBox="0 0 18 18" fill="none">
                                <path d="M7.6 7.6a2 2 0 1 0 2.8 2.8M12.5 12.5A6.5 6.5 0 0 1 9 14c-4.5 0-7-5-7-5a11.5 11.5 0 0 1 3-3.5m2.2-1.2A5.5 5.5 0 0 1 9 4c4.5 0 7 5 7 5a11.5 11.5 0 0 1-1.2 1.8" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
                                <path d="M2 2l14 14" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
                            </svg>
                        </button>
                    </div>
                </div>

                <div class="btn-group">
                    <button type="button" id="testBtn" class="btn-secondary">Test Connection</button>
                    <button type="submit" id="submitBtn" class="btn-primary">Save & Connect</button>
                </div>

                <div id="status" class="status"></div>
            </form>
        </div>

        <a href="https://github.com/salmonumbrella/deputy-cli" target="_blank" class="github-link">
            <svg viewBox="0 0 16 16" fill="currentColor">
                <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
            </svg>
            View on GitHub
        </a>
    </div>

    <script>
        const form = document.getElementById('setupForm');
        const testBtn = document.getElementById('testBtn');
        const submitBtn = document.getElementById('submitBtn');
        const status = document.getElementById('status');
        const togglePassword = document.getElementById('togglePassword');
        const tokenInput = document.getElementById('token');
        const eyeIcon = document.getElementById('eyeIcon');
        const eyeOffIcon = document.getElementById('eyeOffIcon');
        const csrfToken = '{{.CSRFToken}}';

        const requiredFields = ['install', 'geo', 'token'];
        const installInput = document.getElementById('install');
        const geoSelect = document.getElementById('geo');
        const apiUrlHint = document.getElementById('apiUrlHint');
        const apiUrlLink = document.getElementById('apiUrlLink');
        const apiUrlText = document.getElementById('apiUrlText');
        let isBusy = false;

        // Update dynamic API URL
        function updateApiUrl() {
            const install = installInput.value.trim().toLowerCase();
            const geo = geoSelect.value;

            if (install && geo) {
                const url = 'https://' + install + '.' + geo + '.deputy.com/exec/devapp/oauth_clients';
                apiUrlText.innerHTML = '<span class="dynamic">' + install + '</span>.' + geo + '.deputy.com/exec/devapp/oauth_clients';
                apiUrlLink.href = url;
                apiUrlLink.classList.add('active');
                apiUrlLink.onclick = null;
                apiUrlHint.classList.add('ready');
            } else {
                apiUrlText.innerHTML = '<span class="placeholder">Enter install name & region above</span>';
                apiUrlLink.href = '#';
                apiUrlLink.classList.remove('active');
                apiUrlLink.onclick = function() { return false; };
                apiUrlHint.classList.remove('ready');
            }
        }

        installInput.addEventListener('input', updateApiUrl);
        geoSelect.addEventListener('change', updateApiUrl);

        // Clear error state when user types
        requiredFields.forEach(id => {
            const el = document.getElementById(id);
            el.addEventListener('input', function() {
                this.classList.remove('error');
            });
            el.addEventListener('change', function() {
                this.classList.remove('error');
            });
        });

        togglePassword.addEventListener('click', () => {
            const isPassword = tokenInput.type === 'password';
            tokenInput.type = isPassword ? 'text' : 'password';
            eyeIcon.style.display = isPassword ? 'none' : 'block';
            eyeOffIcon.style.display = isPassword ? 'block' : 'none';
        });

        function showStatus(type, message) {
            status.className = 'status show ' + type;
            if (type === 'loading') {
                status.innerHTML = '<div class="spinner"></div><span>' + message + '</span>';
            } else {
                const icon = type === 'success'
                    ? '<svg class="status-icon" viewBox="0 0 16 16" fill="none"><path d="M13 5L6.5 11.5L3 8" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>'
                    : '<svg class="status-icon" viewBox="0 0 16 16" fill="none"><path d="M12 4L4 12M4 4L12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>';
                status.innerHTML = icon + '<span>' + message + '</span>';
            }
        }

        function hideStatus() {
            status.className = 'status';
        }

        function validateRequired() {
            let valid = true;
            requiredFields.forEach(id => {
                const input = document.getElementById(id);
                if (!input.value.trim()) {
                    input.classList.add('error');
                    valid = false;
                }
            });
            return valid;
        }

        function getFormData() {
            return {
                install: document.getElementById('install').value.trim().toLowerCase(),
                geo: document.getElementById('geo').value.trim(),
                token: document.getElementById('token').value.trim()
            };
        }

        testBtn.addEventListener('click', async () => {
            if (isBusy) return;
            isBusy = true;
            hideStatus();
            if (!validateRequired()) {
                isBusy = false;
                return;
            }

            const data = getFormData();
            testBtn.disabled = true;
            submitBtn.disabled = true;
            showStatus('loading', 'Testing connection...');
            try {
                const response = await fetch('/validate', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': csrfToken },
                    body: JSON.stringify(data)
                });
                const result = await response.json();
                showStatus(result.success ? 'success' : 'error', result.success ? 'Connection successful!' : result.error);
            } catch (err) {
                showStatus('error', 'Request failed: ' + err.message);
            } finally {
                testBtn.disabled = false;
                submitBtn.disabled = false;
                isBusy = false;
            }
        });

        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            if (isBusy) return;
            isBusy = true;
            hideStatus();
            if (!validateRequired()) {
                isBusy = false;
                return;
            }

            const data = getFormData();
            testBtn.disabled = true;
            submitBtn.disabled = true;
            showStatus('loading', 'Saving credentials...');
            try {
                const response = await fetch('/submit', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': csrfToken },
                    body: JSON.stringify(data)
                });
                const result = await response.json();
                if (result.success) {
                    showStatus('success', 'Credentials saved! Redirecting...');
                    setTimeout(() => { window.location.href = '/success'; }, 600);
                } else {
                    showStatus('error', result.error);
                    testBtn.disabled = false;
                    submitBtn.disabled = false;
                    isBusy = false;
                }
            } catch (err) {
                showStatus('error', 'Request failed: ' + err.message);
                testBtn.disabled = false;
                submitBtn.disabled = false;
                isBusy = false;
            }
        });
    </script>
</body>
</html>`

const successTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
    <title>Connected - Deputy CLI</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Poppins:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg: #F7F8FC;
            --bg-card: #FFFFFF;
            --bg-terminal: #1A1B2E;
            --border: #E2E4EF;
            --text: #1A1B2E;
            --text-secondary: #5C5E7A;
            --text-muted: #9496AD;
            --primary: #0C017B;
            --primary-light: #ECEAFD;
            --accent: #7F52FD;
            --teal: #37CFCD;
            --success: #10B981;
            --success-light: #D1FAE5;
        }

        * { margin: 0; padding: 0; box-sizing: border-box; }
        html { height: 100%; }

        body {
            font-family: 'Poppins', -apple-system, sans-serif;
            background: var(--bg);
            color: var(--text);
            min-height: 100%;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            padding: 2rem 1.5rem 3rem;
            position: relative;
        }

        body::before {
            content: '';
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background:
                linear-gradient(135deg, rgba(16, 185, 129, 0.04) 0%, transparent 50%),
                linear-gradient(225deg, rgba(55, 207, 205, 0.04) 0%, transparent 50%);
            pointer-events: none;
        }

        .container {
            width: 100%;
            max-width: 400px;
            text-align: center;
            position: relative;
            z-index: 1;
        }

        .success-icon {
            width: 64px;
            height: 64px;
            background: linear-gradient(135deg, var(--success-light) 0%, #A7F3D0 100%);
            border-radius: 50%;
            margin: 0 auto 1.25rem;
            display: flex;
            align-items: center;
            justify-content: center;
            animation: scaleIn 0.5s cubic-bezier(0.34, 1.56, 0.64, 1) forwards;
            box-shadow: 0 8px 24px rgba(16, 185, 129, 0.2);
        }

        @keyframes scaleIn {
            from { transform: scale(0); opacity: 0; }
            to { transform: scale(1); opacity: 1; }
        }

        .success-icon svg {
            width: 32px;
            height: 32px;
            color: var(--success);
        }

        h1 {
            font-size: 1.5rem;
            font-weight: 700;
            margin-bottom: 0.25rem;
            letter-spacing: -0.02em;
            animation: fadeUp 0.5s ease 0.15s both;
        }

        .subtitle {
            color: var(--text-secondary);
            font-size: 0.875rem;
            margin-bottom: 1rem;
            animation: fadeUp 0.5s ease 0.2s both;
        }

        @keyframes fadeUp {
            from { opacity: 0; transform: translateY(10px); }
            to { opacity: 1; transform: translateY(0); }
        }

        .account-badge {
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
            background: var(--primary-light);
            color: var(--primary);
            font-size: 0.8125rem;
            font-weight: 600;
            padding: 0.5rem 1rem;
            border-radius: 100px;
            margin-bottom: 1.25rem;
            animation: fadeUp 0.5s ease 0.25s both;
        }

        .account-badge .dot {
            width: 8px;
            height: 8px;
            background: var(--success);
            border-radius: 50%;
            animation: dotPulse 2s ease-in-out infinite;
        }

        @keyframes dotPulse {
            0%, 100% { opacity: 1; transform: scale(1); }
            50% { opacity: 0.6; transform: scale(0.9); }
        }

        .account-badge .region {
            font-size: 0.6875rem;
            font-weight: 500;
            color: var(--accent);
            background: rgba(127, 82, 253, 0.1);
            padding: 0.125rem 0.375rem;
            border-radius: 4px;
            text-transform: uppercase;
        }

        .terminal {
            background: var(--bg-terminal);
            border-radius: 12px;
            overflow: hidden;
            text-align: left;
            animation: fadeUp 0.5s ease 0.3s both;
            box-shadow: 0 8px 32px rgba(26, 27, 46, 0.15);
        }

        .terminal-bar {
            background: #12131F;
            padding: 0.625rem 0.875rem;
            display: flex;
            align-items: center;
            gap: 0.375rem;
        }

        .terminal-dot {
            width: 10px;
            height: 10px;
            border-radius: 50%;
        }

        .terminal-dot.red { background: #FF5F57; }
        .terminal-dot.yellow { background: #FEBC2E; }
        .terminal-dot.green { background: #28C840; }

        .terminal-body {
            padding: 1rem;
        }

        .terminal-line {
            display: flex;
            align-items: center;
            gap: 0.5rem;
            font-family: 'JetBrains Mono', monospace;
            font-size: 0.75rem;
            margin-bottom: 0.5rem;
            color: #E5E7EB;
        }

        .terminal-line:last-child { margin-bottom: 0; }
        .terminal-prompt { color: var(--teal); user-select: none; }
        .terminal-cmd { color: var(--accent); }
        .terminal-output {
            color: #9496AD;
            padding-left: 1rem;
            margin-top: -0.25rem;
            margin-bottom: 0.625rem;
            font-size: 0.6875rem;
        }

        .terminal-cursor {
            display: inline-block;
            width: 8px;
            height: 16px;
            background: var(--teal);
            animation: cursorBlink 1.2s step-end infinite;
            margin-left: 2px;
            vertical-align: middle;
        }

        @keyframes cursorBlink {
            0%, 50% { opacity: 1; }
            50.01%, 100% { opacity: 0; }
        }

        .message {
            margin-top: 1.25rem;
            padding: 1rem;
            background: var(--primary-light);
            border: 1px solid rgba(12, 1, 123, 0.08);
            border-radius: 12px;
            text-align: center;
            animation: fadeUp 0.5s ease 0.4s both;
        }

        .message-icon {
            font-size: 1.25rem;
            margin-bottom: 0.25rem;
        }

        .message-title {
            font-weight: 600;
            font-size: 0.875rem;
            margin-bottom: 0.125rem;
            color: var(--text);
        }

        .message-text {
            font-size: 0.75rem;
            color: var(--text-secondary);
            line-height: 1.5;
        }

        .message-text code {
            font-family: 'JetBrains Mono', monospace;
            background: var(--bg-card);
            color: var(--primary);
            padding: 0.125rem 0.375rem;
            border-radius: 4px;
            font-size: 0.6875rem;
        }

        .footer {
            margin-top: 1rem;
            font-size: 0.75rem;
            color: var(--text-muted);
            animation: fadeUp 0.5s ease 0.5s both;
        }

        .github-link {
            margin-top: 1.25rem;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            gap: 0.5rem;
            text-decoration: none;
            color: var(--text-muted);
            font-size: 0.75rem;
            font-weight: 500;
            transition: color 0.2s ease;
            width: 100%;
        }

        .github-link:hover {
            color: var(--text-secondary);
        }

        .github-link svg {
            width: 16px;
            height: 16px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="success-icon">
            <svg viewBox="0 0 32 32" fill="none">
                <path d="M8 16L14 22L24 10" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
        </div>

        <h1>You're all set!</h1>
        <p class="subtitle">Deputy CLI is connected and ready to use</p>

        {{if .Install}}
        <div class="account-badge">
            <span class="dot"></span>
            <span>{{.Install}}</span>
            {{if .Geo}}<span class="region">{{.Geo}}</span>{{end}}
        </div>
        {{end}}

        <div class="terminal">
            <div class="terminal-bar">
                <span class="terminal-dot red"></span>
                <span class="terminal-dot yellow"></span>
                <span class="terminal-dot green"></span>
            </div>
            <div class="terminal-body">
                <div class="terminal-line">
                    <span class="terminal-prompt">$</span>
                    <span class="terminal-cmd">deputy</span>
                    <span>employees list</span>
                </div>
                <div class="terminal-output">Fetching employees...</div>
                <div class="terminal-line">
                    <span class="terminal-prompt">$</span>
                    <span class="terminal-cmd">deputy</span>
                    <span>timesheets list</span>
                </div>
                <div class="terminal-output">Listing recent timesheets...</div>
                <div class="terminal-line">
                    <span class="terminal-prompt">$</span>
                    <span class="terminal-cursor"></span>
                </div>
            </div>
        </div>

        <div class="message">
            <div class="message-icon">&larr;</div>
            <div class="message-title">Return to your terminal</div>
            <div class="message-text">You can close this window and start using the CLI.<br>Try running <code>deputy --help</code> to see all commands.</div>
        </div>

        <p class="footer">This window will close automatically.</p>

        <a href="https://github.com/salmonumbrella/deputy-cli" target="_blank" class="github-link">
            <svg viewBox="0 0 16 16" fill="currentColor">
                <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
            </svg>
            View on GitHub
        </a>
    </div>

    <script>fetch('/complete', { method: 'POST', headers: { 'X-CSRF-Token': '{{.CSRFToken}}' } }).catch(() => {});</script>
</body>
</html>`
