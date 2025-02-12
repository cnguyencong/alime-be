const fs = require("fs");
const { createInstance } = require("./index.js");
const { jsonToVideo } = require("./video-parallel.js");

async function run() {
  // create working instance
  // const instance = await createInstance({
  //   key: "nFA5H9elEytDyPyvKL7T",
  // });

  // load sample json
  const json = JSON.parse(fs.readFileSync("./test-data/video.json"));
  //   const page = await instance.createPage();

  // await jsonToVideo(instance, json, { out: "out.mp4" });
  await jsonToVideo(
    () =>
      createInstance({
          key: "nFA5H9elEytDyPyvKL7T",
      }),
    json,
    { out: 'out.mp4', parallel: 1, fps: 30 }
  );
  console.timeEnd('render');
  process.exit(0);

  // await instance.close();
  // process.exit(0);
}

run().catch((e) => console.error(e));
