import { BoundingBox  } from './boundingbox';
import { ImageResult  } from './imageresult';


export class Annotatedimage {
  image: string;
  url: string;
  boxes: BoundingBox[];

}

