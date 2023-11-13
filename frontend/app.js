// ------------------------- //

const UPLOAD_CHUNK_API = "http://127.0.0.1:3000/api/upload/chunk";
const VIDEO_INITIALIZATION_API = "http://127.0.0.1:3000/api/upload/init";
const MAIN_VIDEO_UPLOAD_API = "http://127.0.0.1:3000/api/upload";
const JWT_VALIDATION_API = "http://127.0.0.1:3000/api/validate-jwt";
var userId;

function getCookie() {
  var nameEQ = "jwtToken" + "=";
  var ca = document.cookie.split(";");
  for (var i = 0; i < ca.length; i++) {
    var c = ca[i];
    while (c.charAt(0) == " ") c = c.substring(1, c.length);
    if (c.indexOf(nameEQ) == 0) return c.substring(nameEQ.length, c.length);
  }
  return null;
}

async function checkJWTValidation() {
  try {
    JwtToken = getCookie();
    const response = await fetch(JWT_VALIDATION_API, {
      method: "GET",
      headers: {
        Authorization: "Bearer " + JwtToken,
      },
    });

    if (response.ok) {
      console.log("JWT validation successful");
      const data = await response.json();
      return data.userId;
    } else {
      // Redirect to login
      console.error("JWT validation failed");
      window.location.href = "login.html";
      throw new Error("JWT validation failed");
    }
  } catch (error) {
    console.error("Error during JWT validation:", error);
    throw error;
  }
}
checkJWTValidation().then((resultUserId) => {
  userId = resultUserId;
});

function getChunkData(startByte, endByte, file) {
  if (startByte === 0 && endByte === file.size) {
    console.log(file.name);
    const reader = new FileReader();
    return new Promise((resolve, reject) => {
      reader.onload = () => resolve(reader.result);
      reader.onerror = reject;
      reader.readAsArrayBuffer(file);
    });
  } else {
    const chunk = file.slice(startByte, endByte);
    const reader = new FileReader();
    return new Promise((resolve, reject) => {
      reader.onload = () => resolve(reader.result);
      reader.onerror = reject;
      reader.readAsArrayBuffer(chunk);
    });
  }
}

async function calculateSHA256(data) {
  const hashBuffer = await crypto.subtle.digest("SHA-256", data);
  const hashArray = Array.from(new Uint8Array(hashBuffer));
  const sha256sum = hashArray
    .map((byte) => byte.toString(16).padStart(2, "0"))
    .join("");
  // console.log(sha256sum);
  return sha256sum;
}
async function getChunkFileNamesAndCheckSum(chunkSizeInBytes) {
  const chunkFileNamesInOrder = [];
  const chunkFileCheckSum = {};
  let originalFileCheckSum = "";
  const videoInput = document.getElementById("videoInput");

  if (videoInput.files.length > 0) {
    const file = videoInput.files[0];
    const numberOfChunks = Math.ceil(file.size / chunkSizeInBytes);

    const originalFileData = await getChunkData(0, file.size, file);
    originalFileCheckSum = await calculateSHA256(originalFileData);

    console.log("originalFileCheckSum", originalFileCheckSum);

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
    originalFileCheckSum,
  };
}

async function initializeVideoUploader(
  chunkFileNamesInOrder,
  chunkFileCheckSum,
  originalFileCheckSum
) {
  try {
    const uploadInitializationData = {
      video_id: crypto.randomUUID(),
      user_id: "srao0",
      Status: "",
      chunk_filenames: chunkFileNamesInOrder,
      check_sum_map: chunkFileCheckSum,
      original_file_checksum: originalFileCheckSum,
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

    // return response;
    return { uploadInitializationData, response };
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
      return result;
    }
  };

  return await uploadChunk();
}

async function uploadAllChunks(chunkFileNamesInOrder) {
  const successfulUploads = [];

  for (const chunkFilename of chunkFileNamesInOrder) {
    try {
      const result = await uploadChunkFile(
        5 * 1024 * 1024,
        chunkFilename,
        chunkFileNamesInOrder
      );
      successfulUploads.push(result);
    } catch (error) {
      console.error(error);
      // Handle errors, such as retrying the upload or notifying the user
      return false;
    }
  }

  // Check if all chunks were successfully uploaded
  if (successfulUploads.length === chunkFileNamesInOrder.length) {
    console.log("All chunks uploaded successfully!");

    return true;
  }
  return false;
}

async function mergeChunksAndCleanUp(videoId) {
  data = { video_id: videoId };
  const response = await fetch(MAIN_VIDEO_UPLOAD_API, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  });
  if (response.ok) {
    console.log(await response.json());
    console.log("Video has been merged");
  }
}

async function chunkAndUpload() {
  const CHUNK_SIZE = 5 * 1024 * 1024; //5MB
  const { chunkFileNamesInOrder, chunkFileCheckSum, originalFileCheckSum } =
    await getChunkFileNamesAndCheckSum(CHUNK_SIZE);

  data = await initializeVideoUploader(
    chunkFileNamesInOrder,
    chunkFileCheckSum,
    originalFileCheckSum
  );
  if (data.response.ok) {
    isAllChunksUploaded = await uploadAllChunks(chunkFileNamesInOrder);
  }
  if (isAllChunksUploaded) {
    const videoId = data.uploadInitializationData["video_id"];
    console.log("video id: ", videoId);

    response = await mergeChunksAndCleanUp(videoId);
  }
}
