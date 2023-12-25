// ------------------------- //
const BASE_API_URL = "http://127.0.0.1:3000/api";
// const BASE_API_URL = "http://3.26.147.0:3000/api";

const UPLOAD_CHUNK_API = `${BASE_API_URL}/upload/chunk`;
const VIDEO_INITIALIZATION_API = `${BASE_API_URL}/upload/init`;
const MERGE_CHUNKS_API = `${BASE_API_URL}/upload/merge`;
const JWT_VALIDATION_API = `${BASE_API_URL}/validate-jwt`;
const SIMPLE_UPLOAD_FILE_API = `${BASE_API_URL}/upload/simple-upload`;

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
    console.log("number of chunks: ", numberOfChunks);

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
    jwtToken = getCookie();

    const uploadInitializationData = {
      video_id: crypto.randomUUID(),
      user_id: userId,
      Status: "",
      chunk_filenames: chunkFileNamesInOrder,
      check_sum_map: chunkFileCheckSum,
      original_file_checksum: originalFileCheckSum,
    };

    const response = await fetch(VIDEO_INITIALIZATION_API, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${jwtToken}`,
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
  jwtToken = getCookie();
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
        headers: {
          Authorization: `Bearer ${jwtToken}`,
        },
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

const chunkProgressMap = {};

// async function uploadChunkFile(
//   chunkSizeInBytes,
//   chunkFilename,
//   chunkFileNamesInOrder
// ) {
//   jwtToken = getCookie();
//   const videoInput = document.getElementById("videoInput");
//   const file = videoInput.files[0];

//   const getChunkData = (chunkFilename) => {
//     return new Promise((resolve, reject) => {
//       const chunkIndex = chunkFileNamesInOrder.indexOf(chunkFilename);
//       const chunkStart = chunkIndex * chunkSizeInBytes;
//       const chunkEnd = Math.min((chunkIndex + 1) * chunkSizeInBytes, file.size);
//       const chunk = file.slice(chunkStart, chunkEnd);

//       const reader = new FileReader();
//       reader.onload = (event) => {
//         const arrayBuffer = event.target.result;
//         const blob = new Blob([arrayBuffer], { type: file.type });
//         resolve(blob);
//       };
//       reader.onerror = reject;
//       reader.readAsArrayBuffer(chunk);
//     });
//   };

//   const uploadChunk = () => {
//     return new Promise(async (resolve, reject) => {
//       if (file) {
//         const xhr = new XMLHttpRequest();
//         const formData = new FormData();
//         formData.append(
//           "chunkFile",
//           await getChunkData(chunkFilename),
//           chunkFilename
//         );
//         formData.append("chunkID", chunkFilename);

//         xhr.open("POST", UPLOAD_CHUNK_API, true);
//         xhr.setRequestHeader("Authorization", `Bearer ${jwtToken}`);

//         // // Set up the progress event listener
//         // xhr.upload.addEventListener("progress", (event) => {
//         //   if (event.lengthComputable) {
//         //     const percentage = (event.loaded / event.total) * 100;
//         //     chunkProgressMap[chunkFilename] = percentage;
//         //     console.log(
//         //       `Chunk ${chunkFilename} - ${percentage.toFixed(2)}% uploaded`
//         //     );
//         //   }
//         // });

//         xhr.onreadystatechange = () => {
//           if (xhr.readyState === XMLHttpRequest.DONE) {
//             if (xhr.status === 200) {
//               const result = JSON.parse(xhr.responseText);
//               console.log(`Uploaded chunk ${chunkFilename}`, result);
//               resolve(result);
//             } else {
//               reject(
//                 new Error(`HTTP error: unable to upload chunk ${chunkFilename}`)
//               );
//             }
//           }
//         };

//         // Send the chunk data
//         xhr.send(formData);
//       }
//     });
//   };

//   return uploadChunk();
// }

async function uploadAllChunks(chunkFileNamesInOrder, chunkSizeInBytes) {
  const promises = [];

  for (const chunkFilename of chunkFileNamesInOrder) {
    const promise = uploadChunkFile(
      chunkSizeInBytes,
      chunkFilename,
      chunkFileNamesInOrder
    )
      .then((result) => result)
      .catch((error) => {
        console.error(error);
        throw error;
      });

    promises.push(promise);
  }

  try {
    const results = await Promise.all(promises);

    // Check if all chunks were successfully uploaded
    if (results.length === chunkFileNamesInOrder.length) {
      console.log("All chunks uploaded successfully!");
      return true;
    } else {
      return false;
    }
  } catch (error) {
    // Handle any error that occurred during parallel execution
    console.error("Error during parallel upload:", error);
    return false;
  }
}

async function mergeChunksAndCleanUp(videoId) {
  data = { video_id: videoId };
  const response = await fetch(MERGE_CHUNKS_API, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${jwtToken}`,
    },
    body: JSON.stringify(data),
  });
  if (response.ok) {
    console.log(await response.json());
    console.log("Video has been merged");
  }
}

async function chunkAndUpload() {
  const CHUNK_SIZE = 10 * 1024 * 1024; //10MB
  const { chunkFileNamesInOrder, chunkFileCheckSum, originalFileCheckSum } =
    await getChunkFileNamesAndCheckSum(CHUNK_SIZE);

  data = await initializeVideoUploader(
    chunkFileNamesInOrder,
    chunkFileCheckSum,
    originalFileCheckSum
  );
  console.log("chunkFileNames: ", chunkFileNamesInOrder);
  if (data.response.ok) {
    isAllChunksUploaded = await uploadAllChunks(
      chunkFileNamesInOrder,
      CHUNK_SIZE
    );
  }
  if (isAllChunksUploaded) {
    const videoId = data.uploadInitializationData["video_id"];
    console.log("video id: ", videoId);

    response = await mergeChunksAndCleanUp(videoId);
  }
}

async function simpleUploadFile() {
  var formData = new FormData();
  formData.append("file", document.getElementById("videoInput").files[0]);

  fetch(SIMPLE_UPLOAD_FILE_API, {
    method: "POST",
    body: formData,
  })
    .then((response) => response.json())
    .then((data) => {
      console.log(data.message);
    })
    .catch((error) => {
      console.error("Error:", error);
    });
}
