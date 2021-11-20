var index = -1;
var pos = -1;
var volume = 70;
var dir = '';
var countdown = 0;

function blink(btn) {
    btn.style.color = 'lightGreen';
    sleep(200).then(() => {
        btn.style.color = '';
    });
}

function playPrev() {
    index--;
    pos = 0;
    if (index < 0) {
        index = fileList.length - 1;
    }
    blink(btnPrev);
    playIndex(index, false);
}

function playNext() {
    index++;
    pos = 0;
    if (index === fileList.length) {
        index = 0;
    }
    blink(btnNext);
    playIndex(index, false);
}

function playIndex(index, isInit) {
    title.innerText = unescape(decodeURI(fileList[index]));
    audio.src = "/audio/" + fileList[index];
    audio.volume = volume / 100;
    audio.currentTime = pos;
    btnStatus.innerText = volume;
    if (!isInit) {
        audio.play();
        localStorage.setItem(dir + '_index', index);
    }
}

function sleep (time) {
    return new Promise((resolve) => setTimeout(resolve, time));
}

function volumeUp() {
    volume += 10;
    if (volume > 100) {
        volume = 100;
    }
    audio.volume = volume / 100;
    btnStatus.innerText = volume;
    blink(btnStatus);
    blink(btnVolumeUp);
}

function volumeDown() {
    volume -= 10;
    if (volume < 10) {
        volume = 10;
    }
    audio.volume = volume / 100;
    btnStatus.innerText = volume;
    blink(btnStatus);
    blink(btnVolumeDown);
}

function seekForward() {
    blink(btnForward);
    pos += 30;
    var dur = parseInt(audio.duration.toFixed());
    if (pos >= dur - 2) {
        pos = dur - 2;
    }
    audio.currentTime = pos;
}

function seekBack() {
    blink(btnBack);
    pos -= 30;
    if (pos < 0) {
        pos = 0;
    }
    audio.currentTime = pos;
}

function listReceived(data, isInit){
    fileList = data;
    if (index == null || index >= fileList.length) {
        index = -1;
    }
    pos = localStorage.getItem(dir + '_position');
    if (pos == null) {
        pos = 5;
    } else {
        pos = parseInt(pos);
    }
    pos -= 5;
    if (pos < 0) {
        pos = 0;
    }
    btnStatus.innerText = pos + '/' + audio.duration.toFixed();
    if (index > -1) {
        btnPause.style.color = 'pink';
        btnPlay.style.color = 'lightGreen';
        playIndex(index, isInit);
    } else {
        title.innerText = dir + ' ' + fileList.length + ' files'
    }
}

function openDirPane() {
    mainPane.style.display = 'none';
    dirPane.style.display = '';
}

function openMainPane() {
    dirPane.style.display = 'none';
    listPane.style.display = 'none';
    mainPane.style.display = '';
}

function selectDir(name) {
    localStorage.setItem('dir', name)
    dir = name;
    openMainPane();
    index = localStorage.getItem(dir + '_index');
    if (index == null) {
        index = 0;
    }
    fetch("/files?dir="+dir).then(response => response.json()).then(data => listReceived(data, false));
}

function addTimer(seconds, btn) {
    blink(btn);
    countdown += seconds;
}

var oldpos = 0;

function init() {
    audio.addEventListener('pause', (event) => {
        btnPause.style.color = 'pink';
        btnPlay.style.color = 'lightGreen';
    });
    audio.addEventListener('play', (event) => {
        btnPause.style.color = 'lightGreen';
        btnPlay.style.color = 'pink';
    });

    audio.addEventListener('timeupdate', (event) => {
        var newpos = parseInt(audio.currentTime.toFixed());
        if (newpos != pos) {
            if (countdown > 0) {
                countdown--;
                if (countdown == 0) {
                    audio.pause();
                }
            }

            pos = newpos;
            var msg = pos + '/' + audio.duration.toFixed();
            if (countdown > 0) {
                msg += '//' + (countdown/60).toFixed();
            }

            btnStatus.innerText = msg;
            localStorage.setItem(dir + '_position', pos);
        }
    });
    audio.addEventListener('ended', (event) => {
        playNext();
    });

    index = -1;
    
    dir = localStorage.getItem('dir');
    if (dir != null) {
        index = localStorage.getItem(dir + '_index');
        if (index == null) {
            index = -1;
        } else {
            index = parseInt(index);
        }
        fetch("/files?dir="+dir).then(response => response.json()).then(data => listReceived(data, true));
    }
}

function openListPane() {
    mainPane.style.display = 'none';
    listPane.style.display = '';
}

function fillList(data) {
    songList.innerHTML = data;
}

function showList() {
    openListPane();
    fetch("/songlist?dir="+dir).then(response=>response.text()).then(t => fillList(t));
}
