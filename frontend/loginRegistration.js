const BASE_API_URL = "http://127.0.0.1:3000/api";
// const BASE_API_URL = "http://3.26.147.0:3000/api";

const LOGIN_API = `${BASE_API_URL}/login/userid-password`;
const REGISTER_SEND_EMAIL_OTP_API = `${BASE_API_URL}/register/send-email-otp`;
const VERIFY_OTP_API = `${BASE_API_URL}/register/verify-otp`;
const CREATE_NEW_USER_API = `${BASE_API_URL}/register/create-user`;

async function handleLoginResponse(response) {
  if (response.ok) {
    console.log("Login successful");
    responseJson = await response.json();

    setCookie("jwtToken", responseJson.token, 1);
    window.location.href = "dashboard.html";
  } else {
    console.error("Login failed:", response.message);
    alert("Invalid username or password. Please try again.");
  }
}

function setCookie(name, value, days) {
  var expires = "";
  if (days) {
    var date = new Date();
    date.setTime(date.getTime() + days * 24 * 60 * 60 * 1000);
    expires = "; expires=" + date.toUTCString();
  }
  document.cookie =
    name + "=" + (value || "") + expires + ";path=/;SameSite=None;Secure;";
}

function getCookie() {
  console.log("cookie");
  var nameEQ = "jwtToken" + "=";
  var ca = document.cookie.split(";");
  for (var i = 0; i < ca.length; i++) {
    var c = ca[i];
    while (c.charAt(0) == " ") c = c.substring(1, c.length);
    if (c.indexOf(nameEQ) == 0) {
      console.log(c.substring(nameEQ.length, c.length));
      return c.substring(nameEQ.length, c.length);
    }
  }

  return null;
}

function clearCookie(name = "jwtToken") {
  document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 UTC;path=/;SameSite=None;Secure;`;
}

function loginUser() {
  const userId = document.getElementById("userid").value;
  const password = document.getElementById("password").value;

  clearCookie("jwtToken");

  fetch(LOGIN_API, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ userId: userId, password: password }),
  })
    .then((response) => handleLoginResponse(response))
    .catch((error) => console.error("Login request failed:", error));
}

function sendOTP() {
  const userId = document.getElementById("new-userid").value;
  const email = document.getElementById("new-email").value;

  document.getElementById("otp-section").style.display = "block";
  document.getElementById("send-otp-button").disabled = true;

  fetch(REGISTER_SEND_EMAIL_OTP_API, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ userEmail: email, userId: userId }),
  })
    .then((response) => {
      if (!response.ok) {
        throw new Error(`Error sending OTP: ${response.status}`);
      }
      document.getElementById("new-userid").readOnly = true;
      document.getElementById("new-email").readOnly = true;
      document.getElementById("new-userid").classList.add("readonly");
      document.getElementById("new-email").classList.add("readonly");
    })
    .catch((error) => {
      console.error("Error:", error.message);
    });
}

function verifyOTP() {
  const userId = document.getElementById("new-userid").value;
  const email = document.getElementById("new-email").value;
  const enteredOTP = document.getElementById("otp").value;

  fetch(VERIFY_OTP_API, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ userId: userId, userEmail: email, OTP: enteredOTP }),
  })
    .then((response) => {
      if (!response.ok) {
        throw new Error(`Error verifying OTP: ${response.status}`);
      }
      return response.json();
    })
    .then((verifyResponse) => {
      console.log("OTP verified successfully!");

      document.getElementById("otp-section").style.display = "none";
      document.getElementById("username-password-section").style.display =
        "block";
    })
    .catch((error) => {
      console.error("Error:", error.message);
    });
}

function submitRegistrationDetails() {
  const userId = document.getElementById("new-userid").value;
  const email = document.getElementById("new-email").value;
  const username = document.getElementById("new-username").value;
  const password = document.getElementById("new-password").value;

  fetch(CREATE_NEW_USER_API, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      userId: userId,
      userEmail: email,
      userName: username,
      password: password,
    }),
  })
    .then((response) => {
      console.log(response);
      if (!response.ok) {
        throw new Error(`Error creating user: ${response.status}`);
      }
      return response.json();
    })
    .then((data) => {
      console.log("User created successfully!");
      window.location.href = "login.html";
    })
    .catch((error) => {
      console.error("Error:", error.message);
    });
}
