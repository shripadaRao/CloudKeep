function login() {
  const username = document.getElementById("username").value;
  const password = document.getElementById("password").value;

  // Perform user authentication (e.g., check credentials against a database)
  if (authenticateUser(username, password)) {
    // Redirect the user to a secure dashboard or other page
    window.location.href = "dashboard.html";
  } else {
    alert("Invalid username or password. Please try again.");
  }
}

function authenticateUser(username, password) {
  // Implement secure authentication logic here
  // In a real application, you should hash and salt the password and check it against a database of users.
  // For enhanced security, consider using a secure authentication library.

  // In this basic example, we'll check against hard-coded credentials for demonstration purposes.
  const validUsername = "user123";
  const validPassword = "password123";

  return username === validUsername && password === validPassword;
}

function registerUser() {
  const newUsername = document.getElementById("new-username").value;
  const newPassword = document.getElementById("new-password").value;

  // Basic client-side input validation (you should perform more robust validation on the server)
  if (newUsername.trim() === "" || newPassword.trim() === "") {
    alert("Please fill out both fields.");
    return;
  }
}
const UPLOAD_CHUNK_API = "http://127.0.0.1:3000/api/upload/chunk";
const VIDEO_INITIALIZATION_API = "http://127.0.0.1:3000/api/upload/init";

function getChunkData(startByte, endByte, file) {
  const chunk = file.slice(startByte, endByte);
  const reader = new FileReader();
  return new Promise((resolve, reject) => {
    reader.onload = () => resolve(reader.result);
    reader.onerror = reject;
    reader.readAsArrayBuffer(chunk);
  });
}

async function calculateSHA256(data) {
  // const buffer = new TextEncoder().encode(data);
  const hashBuffer = await crypto.subtle.digest("SHA-256", data);
  const hashArray = Array.from(new Uint8Array(hashBuffer));
  const sha256sum = hashArray
    .map((byte) => byte.toString(16).padStart(2, "0"))
    .join("");
  console.log(sha256sum);
  return sha256sum;
}
async function getChunkFileNamesAndCheckSum(chunkSizeInBytes) {
  const chunkFileNamesInOrder = [];
  const chunkFileCheckSum = {};
  const videoInput = document.getElementById("videoInput");

  if (videoInput.files.length > 0) {
    const file = videoInput.files[0];
    const numberOfChunks = Math.ceil(file.size / chunkSizeInBytes);

    for (let i = 0; i < numberOfChunks; i++) {
      const startByte = i * chunkSizeInBytes;
      const endByte = Math.min((i + 1) * chunkSizeInBytes, file.size);
      const chunkFileName = crypto.randomUUID();
      chunkFileNamesInOrder.push(chunkFileName);
      const chunkData = await getChunkData(startByte, endByte, file);
      const sha256sum = await calculateSHA256(chunkData);
      chunkFileCheckSum[chunkFileName] = sha256sum;
    }
  }
  return {
    chunkFileNamesInOrder,
    chunkFileCheckSum,
  };
}

async function initializeVideoUploader(
  chunkFileNamesInOrder,
  chunkFileCheckSum
) {
  try {
    const uploadInitializationData = {
      video_id: crypto.randomUUID(),
      user_id: "srao0",
      Status: "",
      chunk_filenames: chunkFileNamesInOrder,
      check_sum_map: chunkFileCheckSum,
    };

    const response = await fetch(VIDEO_INITIALIZATION_API, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(uploadInitializationData),
    });

    const data = await response.json();
    console.log("Success:", data);

    return response;
  } catch (error) {
    console.error("Error:", error);
    throw error;
  }
}

async function uploadChunkFile(
  chunkSizeInBytes,
  chunkFilename,
  chunkFileNamesInOrder
) {
  const videoInput = document.getElementById("videoInput");
  const file = videoInput.files[0];

  const getChunkData = (chunkFilename) => {
    return new Promise((resolve, reject) => {
      const chunkIndex = chunkFileNamesInOrder.indexOf(chunkFilename);
      const chunkStart = chunkIndex * chunkSizeInBytes;
      const chunkEnd = Math.min((chunkIndex + 1) * chunkSizeInBytes, file.size);
      const chunk = file.slice(chunkStart, chunkEnd);

      const reader = new FileReader();
      reader.onload = (event) => {
        const arrayBuffer = event.target.result;
        const blob = new Blob([arrayBuffer], { type: file.type });
        resolve(blob);
      };
      reader.onerror = reject;
      reader.readAsArrayBuffer(chunk);
    });
  };

  const uploadChunk = async () => {
    if (file) {
      const formData = new FormData();
      formData.append(
        "chunkFile",
        await getChunkData(chunkFilename),
        chunkFilename
      );
      formData.append("chunkID", chunkFilename);

      const response = await fetch(UPLOAD_CHUNK_API, {
        method: "POST",
        body: formData,
      });

      if (!response.ok) {
        throw new Error(`HTTP error: unable to upload chunk ${chunkFilename}`);
      }

      const result = await response.json();
      console.log(`uploaded chunk ${chunkFilename}`, result);
    }
  };

  await uploadChunk();
}

async function uploadAllChunks(chunkFileNamesInOrder) {
  for (const chunkFilename of chunkFileNamesInOrder) {
    await uploadChunkFile(
      5 * 1024 * 1024,
      chunkFilename,
      chunkFileNamesInOrder
    );
  }
}

async function chunkAndUpload() {
  const CHUNK_SIZE = 5 * 1024 * 1024; //5MB
  const { chunkFileNamesInOrder, chunkFileCheckSum } =
    await getChunkFileNamesAndCheckSum(CHUNK_SIZE);

  response = await initializeVideoUploader(
    chunkFileNamesInOrder,
    chunkFileCheckSum
  );
  if (response.ok) {
    uploadAllChunks(chunkFileNamesInOrder);
  }
}
