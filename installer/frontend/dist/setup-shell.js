const $ = (id) => document.getElementById(id)

// appName is the product's name, read from the setup program rather than
// written here. The page renders before the name arrives, so the static text
// carries none of it and every screen fills it in from this. Written down here
// it would go stale through a rename without a word from anything.
let appName = ''

// currentState is the one reading of the machine, kept so a screen that goes
// back can return to the one that was due.
let currentState = null

function backend() {
    return window.go && window.go.main && window.go.main.App
}

/* ------------------------------------------------------------------ theme */

// applyTheme sets the theme and points the button at the theme it would switch to
// (NFR-UX-004), so the sun shows while the page is dark. Repainting and re-facing
// the toggle happen here together. The two icons are the application's own
// artwork, written in beside this page by tools/genicons.py.
function applyTheme(theme) {
    document.documentElement.setAttribute('data-theme', theme)
    $('theme-icon').src = theme === 'dark' ? 'light-mode.png' : 'dark-mode.png'
    $('theme').title = theme === 'dark' ? 'Switch to light' : 'Switch to dark'
    $('theme').setAttribute('aria-label', $('theme').title)
}

function currentTheme() {
    return document.documentElement.getAttribute('data-theme') === 'dark' ? 'dark' : 'light'
}

$('theme').onclick = () => applyTheme(currentTheme() === 'dark' ? 'light' : 'dark')

/* ---------------------------------------------------------------- screens */

// shownScreen and shownFooter are what is on screen now, so the licence can be
// read from anywhere and then hand the same screen back.
let shownScreen = ''
let shownFooter = []

function showScreen(name) {
    document.querySelectorAll('.screen').forEach((el) => el.classList.remove('active'))
    $('screen-' + name).classList.add('active')
    shownScreen = name
}

function setFooter(buttons) {
    shownFooter = buttons
    const footer = $('footer')
    footer.innerHTML = ''
    buttons.forEach((spec) => {
        const el = document.createElement('button')
        el.className = 'btn' + (spec.kind ? ' ' + spec.kind : '')
        el.textContent = spec.label
        el.onclick = (ev) => { acted = true; spec.onClick(ev) }
        footer.appendChild(el)
    })
    if (acted) focusFooter()
}

// acted is false until the reader first presses a footer action. Until then the
// window keeps a neutral start (installer skill): nothing wears a ring until the
// first Tab. After it, each new screen leads with its own go-ahead.
let acted = false

// focusFooter puts focus on the button a screen leads with, so Enter does the
// obvious thing and the ring says where it would land.
function focusFooter() {
    const footer = $('footer')
    const first = footer.querySelector('.btn.primary') || footer.querySelector('.btn')
    if (first) first.focus()
}


/* ---------------------------------------------------------------- licence */

// licenceFile is the application's own LICENSE, copied in beside this page by build.ps1
// so the text has one home at the repository root.
const licenceFile = 'LICENSE.txt'

// showLicence is a screen like any other. Back hands over exactly the screen and
// footer that were showing, so reading it never changes what setup was doing.
async function showLicence() {
    if (shownScreen === 'licence') return
    const back = {screen: shownScreen, footer: shownFooter}
    const text = $('licence-text')
    try {
        const response = await fetch(licenceFile)
        if (!response.ok) throw new Error(String(response.status))
        text.textContent = await response.text()
    } catch (e) {
        text.textContent = 'The licence text is not in this build of setup.'
    }
    showScreen('licence')
    // The licence reads itself (scroll skill); a fresh cycle each time it opens.
    text.scrollTop = 0
    const stopReading = startAutoScroll(text)
    setFooter([{
        label: 'Back', kind: 'primary',
        onClick: () => { stopReading(); showScreen(back.screen); setFooter(back.footer) },
    }])
}

$('licence').onclick = () => { void showLicence() }

/* ---------------------------------------------------------------- options */

// renderOptions fills a container with checkboxes and returns a reader for their
// values, so no screen has to know the ids of its own boxes.
function renderOptions(container, specs) {
    container.innerHTML = ''
    const boxes = {}
    specs.forEach((spec) => {
        const label = document.createElement('label')
        label.className = 'option'
        const input = document.createElement('input')
        input.type = 'checkbox'
        input.checked = !!spec.checked
        if (spec.onChange) input.onchange = () => spec.onChange(input.checked)
        const tick = document.createElement('span')
        tick.className = 'check'
        const text = document.createElement('span')
        const title = document.createElement('span')
        title.className = 'label'
        title.textContent = spec.label
        text.appendChild(title)
        if (spec.hint) {
            const hint = document.createElement('span')
            hint.className = 'hint'
            hint.textContent = spec.hint
            text.appendChild(hint)
        }
        label.append(input, tick, text)
        container.appendChild(label)
        boxes[spec.key] = input
    })
    return (key) => boxes[key].checked
}

// freshChoices are what a first install applies. Reinstall reuses them, which is
// the whole of the difference from a repair: a repair leaves every choice alone,
// a reinstall puts the install back to how a new one would look.
const freshChoices = {startMenu: true, desktop: true}

// launchOption finishes every screen that writes files. Setup's job is done once
// the application is running, so the same tick that starts it also closes setup.
// Checked by default, because that is why people run an installer.
function launchOption() { return {
    key: 'launch',
    label: `Start ${appName} and close setup when this finishes`,
    checked: true,
} }

// shortcutOptions are the two boxes every screen that offers choices carries.
// They open on what is already true of the machine once it is installed.
function shortcutOptions(state) {
    return [
        {
            key: 'startMenu', label: 'Add a Start Menu entry',
            hint: 'Find it by typing its name in the Start Menu.',
            checked: state.installed ? state.startMenu : true,
        },
        {
            key: 'desktop', label: 'Add a Desktop shortcut',
            checked: state.installed ? state.desktop : true,
        },
    ]
}

/* ------------------------------------------------------------------- work */

function onProgress(p) {
    $('progress-fill').style.width = p.pct + '%'
    $('progress-status').textContent = p.msg
}

async function run(work, title, doneTitle, doneMsg) {
    $('progress-title').textContent = title
    $('progress-fill').style.width = '0'
    $('progress-status').textContent = 'Starting...'
    setFooter([])
    showScreen('progress')
    try {
        await work()
        $('done-title').textContent = doneTitle
        $('done-msg').textContent = doneMsg
        showScreen('done')
        setFooter([{label: 'Close', kind: 'primary', onClick: () => backend().Quit()}])
    } catch (e) {
        showError(String(e))
    }
}

// finish runs one piece of work, then honours the launch tick. A successful
// launch closes setup; a failed one leaves the error on screen, so setup never
// disappears having quietly failed.
function finish(work, wanted, title, doneTitle, doneMsg) {
    return withAppClosed(() => run(
        () => work().then(() => {
            if (wanted) return backend().LaunchApp().then(() => backend().Quit())
        }),
        title, doneTitle, doneMsg + (wanted ? ' It is starting now.' : ''),
    ))
}

// showError is the failure verdict. It names where the step log was written, so
// the one place that says what happened is never a thing to go looking for.
function showError(message) {
    const where = currentState && currentState.logPath
    $('error-msg').textContent = message + (where ? ` The step log is at ${where}` : '')
    showScreen('error')
    setFooter([{label: 'Close', kind: 'primary', onClick: () => backend().Quit()}])
}

// withAppClosed runs the work once the application is not running. If it is, the
// offer to close it comes first, rather than failing later on a locked file.
async function withAppClosed(proceed) {
    if (!(await backend().AppRunning())) {
        proceed()
        return
    }
    showScreen('running')
    setFooter([
        {label: 'Cancel', onClick: () => route(currentState)},
        {
            label: 'Close it and continue', kind: 'primary', onClick: async () => {
                setFooter([])
                try {
                    await backend().CloseRunningApp()
                } catch (e) {
                    showError(String(e))
                    return
                }
                proceed()
            },
        },
    ])
}

// install runs one write of the files, whatever the screen called it. The launch
// option is read here so a successful launch closes setup rather than leaving it
// waiting on a Close button nobody needs.
function install(read, title, doneTitle, doneMsg) {
    const launchAfter = read('launch')
    const choices = {
        startMenu: read('startMenu'),
        desktop: read('desktop'),
    }
    return withAppClosed(() => run(
        () => backend().Install(choices).then(() => {
            if (launchAfter) return backend().LaunchApp().then(() => backend().Quit())
        }),
        title, doneTitle,
        doneMsg + (launchAfter ? ' It is starting now.' : ''),
    ))
}
