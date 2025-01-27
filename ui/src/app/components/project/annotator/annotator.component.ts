import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { ProjectService } from '@app/services/project.service';
import { Project } from '@app/classes/project';
import { Annotator } from '@app/classes/annotator';
import { BoundingBox } from '@app/classes/boundingbox';
import { Annotation } from '@app/classes/annotation';
import { ImageResult } from '@app/classes/imageresult';

import {NgbModal, ModalDismissReasons} from '@ng-bootstrap/ng-bootstrap';
import {Annotatedimage} from '@app/classes/annotatedimage';
import {forEach} from '@angular/router/src/utils/collection';

@Component({
  selector: 'app-annotator',
  templateUrl: './annotator.component.html',
  styleUrls: ['./annotator.component.css']
})
export class AnnotatorComponent implements OnInit {
  annotator: Annotator;
  currentBox: BoundingBox;
  boxes = new Array<BoundingBox>();
  project = new Project();
  image: ImageResult ;

  constructor(
    private route: ActivatedRoute,
    private projectService: ProjectService,
    private modalService: NgbModal,
  ) { }

  ngOnInit() {
    this.route.params.subscribe(param => {
      console.log('Param => ', param);
      this.projectService.getProject(param.name).subscribe(project => {
        this.project = project;
        this.annotator = new Annotator('#annotator');

        this.annotator.boxes.subscribe(box => {
          this.currentBox = box;
        });

        this.annotator.removeBox.subscribe(box => {
          const index: number = this.boxes.indexOf(box);
          if (index !== -1) {
            this.boxes.splice(index, 1);
          }
          this.annotator.setBoundingBoxes(this.boxes);
        });

        this.nextImage();
      });
    });
  }

  setLabel(label: string) {
    if (!this.currentBox) { return; }
    if ((this.currentBox.width - this.currentBox.x) * this.annotator.canvas.clientWidth < 20 ||
        (this.currentBox.height - this.currentBox.y) * this.annotator.canvas.clientHeight < 20
    ) {
      alert('Box too small !');
      return;
    }

    this.currentBox.label = label;
    this.boxes.push(this.currentBox);
    this.annotator.setBoundingBoxes(this.boxes);
    this.currentBox = null;
  }


  saveAnnotation() {
    console.log(this.boxes);
    const annotation = new Annotation();
    annotation.image = this.image.name;
    annotation.boxes = this.boxes;
    console.log(annotation);
    this.projectService.saveAnnotation(this.project, annotation).subscribe(ann => {
      console.log('saved:', ann);
      this.nextImage();
    });
  }

  saveEmptyAnnotation(content) {
    this.modalService.open(content, {}).result.then(
      result => {
        console.log('result', result);
      },
      reason => {
        console.log('reason', reason);
      }
    );
  }

  nextImage() {
    // TODO: send box to server before to get next image
    this.projectService.getNextImage().subscribe(image => {
      console.log(image);
      this.image = new ImageResult(image.image, image.url);
      this.boxes = new Array<BoundingBox>();
      console.log('gonna load img');

      this.annotator.loadImage(image, this);
     /* image.boxes.forEach(box => {
        const mybox = new BoundingBox(box.x, box.y, box.width, box.height);
        this.currentBox = mybox;
        console.log('gonna set Label');
        // this.setLabel(box.label);
         mybox.label = box.label;
         this.boxes.push(mybox);
      });
         this.annotator.setBoundingBoxes(this.boxes);
      // this.annotator.drawBoundingBoxes()
      */
      console.log('the image url : ');
      console.log(this.image.url);
      console.log('the image name : ');
      console.log(this.image.name);
      console.log('the boxes after : ');
      console.log(this.boxes);
    });
  }
}
