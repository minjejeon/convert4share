export function EventsOn(eventName, callback) {
    if (window['runtime']) {
        return window['runtime']['EventsOn'](eventName, callback);
    }
    return () => {};
}
export function EventsOff(eventName) {
    if (window['runtime']) {
        window['runtime']['EventsOff'](eventName);
    }
}
export function EventsEmit(eventName, data) {
    if (window['runtime']) {
        window['runtime']['EventsEmit'](eventName, data);
    }
}
export function WindowSetAlwaysOnTop(b) {
     if (window['runtime']) {
        window['runtime']['WindowSetAlwaysOnTop'](b);
    }
}
export function WindowShow() {
     if (window['runtime']) {
        window['runtime']['WindowShow']();
    }
}
export function WindowUnminimise() {
     if (window['runtime']) {
        window['runtime']['WindowUnminimise']();
    }
}
