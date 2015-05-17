/**
 * Created by gernest on 5/16/15.
 */

$(document).ready(function(){
    var previewTmpl=$('#template'),
        cancelBtn=$('button.cancel'),
        startBtn=$('button.start');
    previewTmpl.detach();
   var  dz=$('.drop-here').dropzone({
        url: "/uploads", // Set the url
        parallelUploads: 20,
        previewTemplate: previewTmpl.html(),
        autoQueue: false, // Make sure the files aren't queued until manually added
        previewsContainer: "#previews", // Define the container to display the previews
        clickable: ".fileinput-button"
    });
    dz.on("addedfile", function(file){
        dz.enqueueFile(file);
    });
    dz.on("totaluploadprogress",function(progress){
        $('#total-progress.progress-bar').css("width",progress+"%");
    });
    dz.on("sending", function(file){
        $('#total-progress').css("opacity",'1');
        startBtn.toggleClass("disabled");
    });
    dz.on("queuecomplete",function(progress){
        $('#total-progress').css("opacity",'0');
    })
    startBtn.click(function(){
       dz.enqueueFiles(dz.getFilesWithStatus(Dropzone.ADDED))
    })
    cancelBtn.click(function(){
        dz.removeAllFiles(true);
    })
})