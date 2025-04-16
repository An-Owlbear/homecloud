import http from 'k6/http';
import { sleep, check } from 'k6';

export const options = {
    duration: '250s',
    vus: 20,
}

const hostname = 'myserver.homecloudapp.com'
const scheme = 'https'

let token = '';

const defaultOptions = {
    headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
    },
}


export function setup() {
    // Assumes the user is as below and logs in
    const flowResponse = http.get(`${scheme}://kratos.${hostname}/self-service/login/browser`, defaultOptions);
    const flowBody = JSON.parse(flowResponse.body);
    const csrfToken = flowBody.ui.nodes[0].attributes.value;

    const loginBody = {
        identifier: 'yetanotheremail@test.com',
        password: 'yetanotherpass',
        method: 'password',
        csrf_token: csrfToken,
    }

    const loginResponse = http.post(`${scheme}://kratos.${hostname}/self-service/login?flow=${flowBody.id}`, JSON.stringify(loginBody), defaultOptions);

    check(loginResponse , {
        "has cookie 'ory_kratos_session'": (r) => r.cookies.ory_kratos_session.length > 0,
    })

    return { token: loginResponse.cookies.ory_kratos_session[0].value };
}

export default function (data) {
    const response = http.get(`${scheme}://${hostname}/api/v1/packages/search`, {
        cookies: { ory_kratos_session: data.token },
    });
    sleep(1);
}