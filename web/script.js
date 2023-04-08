const canvas = document.querySelector('canvas');
const ctx = canvas.getContext('2d');
const imageData = ctx.createImageData(2000, 2000);

const colors = [
  [0x00, 0x00, 0x00],
  [0x00, 0x75, 0x6F],
  [0x00, 0x9E, 0xAA],
  [0x00, 0xA3, 0x68],
  [0x00, 0xCC, 0x78],
  [0x00, 0xCC, 0xC0],
  [0x24, 0x50, 0xA4],
  [0x36, 0x90, 0xEA],
  [0x49, 0x3A, 0xC1],
  [0x51, 0x52, 0x52],
  [0x51, 0xE9, 0xF4],
  [0x6A, 0x5C, 0xFF],
  [0x6D, 0x00, 0x1A],
  [0x6D, 0x48, 0x2F],
  [0x7E, 0xED, 0x56],
  [0x81, 0x1E, 0x9F],
  [0x89, 0x8D, 0x90],
  [0x94, 0xB3, 0xFF],
  [0x9C, 0x69, 0x26],
  [0xB4, 0x4A, 0xC0],
  [0xBE, 0x00, 0x39],
  [0xD4, 0xD7, 0xD9],
  [0xDE, 0x10, 0x7F],
  [0xE4, 0xAB, 0xFF],
  [0xFF, 0x38, 0x81],
  [0xFF, 0x45, 0x00],
  [0xFF, 0x99, 0xAA],
  [0xFF, 0xA8, 0x00],
  [0xFF, 0xB4, 0x70],
  [0xFF, 0xD6, 0x35],
  [0xFF, 0xF8, 0xB8],
  [0xFF, 0xFF, 0xFF],
]

const chunksize = 10;
const splitter = new TransformStream({
  chunkBuffer: new Uint8Array(chunksize),
  chunkBufferLen: 0,
  transform(chunk, controller) {
    chunk = chunk;
    let offset = 0

    if (this.chunkBufferLen > 0) {
      topItOff = chunk.slice(0, chunksize - this.chunkBufferLen)
      offset = chunksize - this.chunkBufferLen
      this.chunkBuffer.set(topItOff, this.chunkBufferLen);
      this.chunkBufferLen = 0;
      controller.enqueue(this.chunkBuffer);
      this.chunkBuffer = new Uint8Array(chunksize);
    }

    while (offset + chunksize <= chunk.length) {
      controller.enqueue(chunk.slice(offset, offset + chunksize));
      offset += chunksize;
    }

    if (offset < chunk.length) {
      const leftovers = chunk.slice(offset);
      this.chunkBufferLen = chunk.length - offset;
      this.chunkBuffer.set(leftovers, 0);
    }
  }
});

let container = document.getElementById("container")
let thing = document.getElementById("zoom")
let j = document.getElementById("move")
let scale = 1;
const scaleFac = 1.2;
let x = 500;
let y = 500;
let mouseX = 0;
let mouseY = 0;
let paused = false;
addEventListener("wheel", (event) => {
  if (!event.deltaY) {
    return;
  }
  const coef = event.deltaY < 0 ? scaleFac : 1 / scaleFac;
  scale *= coef;
  thing.setAttribute("style", `transform: scale(${scale});`);
});
let mouseDown = false;
addEventListener("mousedown", () => {
  mouseDown = true;
});
j.setAttribute("style", `transform: translate(${x}px, ${y}px);`);
addEventListener("mousemove", (move) => {
  mouseX = move.clientX;
  mouseY = move.clientY;
  if (!mouseDown) {
    return;
  }
  x += move.movementX / scale;
  y += move.movementY / scale;
  j.setAttribute("style", `transform: translate(${x}px, ${y}px);`);
});
addEventListener("mouseup", () => {
  mouseDown = false;
});
addEventListener("resize", () => {
  container.setAttribute("style", `width: ${window.innerWidth}; height: ${window.innerHeight};`)
})


async function streamFile(url) {
  const response = await fetch(url);
  const body = response.body;
  const byeah = body.pipeThrough(splitter).getReader();
  const ohyeah = async () => {
    const chunk = (await byeah.read()).value;
    const x1 = (chunk[3] ^ 0x1F) >> 5 | chunk[4] << 3;
    const y1 = chunk[5] | (chunk[6] & 0x7) << 8;
    const x2 = (chunk[6] ^ 0x7) >> 3 | (chunk[7] & 0x3F) << 5;
    const y2 = chunk[8] | (chunk[9] & 0x7) << 8;
    const color = chunk[9] >> 3;
    if (x2 == 0) {
      setColor(x1, y1, color);
    } else {
      x = Math.min(x1, x2);
      y = Math.min(y1, y2);
      xdist = Math.abs(x1-x2);
      ydist = Math.abs(y1-y2);
      for (let i = 0; i++; i<xdist) {
        for (let j = 0; i++; i<ydist) {
          setColor(x+i, y+j, color);
        };
      };
    };
  };
  while (true) {
    if(!paused){
      for (i = 0; i < 2000; i++) {
        await ohyeah();
      }
      ctx.putImageData(imageData, 0, 0);
    };
    let time = await new Promise(resolve => requestAnimationFrame((t) => { resolve(t) }));
    fps.innerHTML = time;
  }
}


function setColor(x, y, color) {
  if(x>=2000 || y>=2000){
    return;
  };
  const index = (y * imageData.width + x) * 4;
  if(index>imageData.data.length){
    console.log(x, y, imageData.data.length);
  };
  imageData.data.set(colors[color], index);
  imageData.data[index + 3] = 255;
}

const pauseplaybtn = document.getElementById("pauseplay");
function pauseplay() {
  if(paused){
    pauseplaybtn.innerHTML = "Pause";
    paused = false;
  }else{
    pauseplaybtn.innerHTML = "Play";
    paused = true;
  }
}

streamFile('packed-sorted');
