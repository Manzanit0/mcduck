import { getCookies } from "$std/http/cookie.ts";
import { decodeJwt } from "$jose/index.ts";

const authCookieName = "_mcduck_fresh_key";

export interface AuthState {
    loggedIn: boolean;
    authToken?: string;
    userEmail?: string;
}

export function getAuthTokenFromBrowser(): AuthState {
    const authToken = getBrowserCookie(authCookieName)
    if (!authToken || authToken === "") {
        return { loggedIn: false }
    }

    return state(authToken)
}

export function getAuthTokenFromRequest(req: Request): AuthState {
    const cookies = getCookies(req.headers);
    const authToken = cookies[authCookieName];
    if (!authToken || authToken === "") {
        return { loggedIn: false }
    }

    return state(authToken)
}

function state(authToken: string): AuthState {
    const loggedIn = true
    const jwt = decodeJwt(authToken);
    const userEmail = jwt.sub!;

    return { loggedIn, authToken, userEmail }
}

function getBrowserCookie(cname: string) {
    const name = cname + "=";
    const decodedCookie = decodeURIComponent(document.cookie);
    const ca = decodedCookie.split(';');
    for (let i = 0; i < ca.length; i++) {
        let c = ca[i];
        while (c.charAt(0) == ' ') {
            c = c.substring(1);
        }
        if (c.indexOf(name) == 0) {
            return c.substring(name.length, c.length);
        }
    }
    return "";
}
