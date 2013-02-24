'use strict';
function PauzerCtrl($scope, $http){

    var master = {
        limit: 100,
        time: 15,
        counterRunning: false,
    },
    intervalVar,
    startCounter;


    $scope.form = angular.copy(master);
    $scope.progressBar = { 
        width: '100%' ,
    };

    $http.get('state').success(function(data){
        if(data.limit > 0){
            $scope.form = {
                limit: data.limit,
                time: data.time,
            };
            startCounter = data.time * 60;
            startTimer(data.secondsLeft);
        }
    });

    $scope.toggleTimer = function(){
        if($scope.counter > 0)
            clearTimer();
        else
            setTimer();
    };

    function setTimer(){
        $http.get('action/' + $scope.form.time + '/' + $scope.form.limit).success(function(data){ 
            startCounter=  $scope.form.time * 60;
            startTimer( startCounter );
        });
    };

    function clearTimer(){
        clearInterval(intervalVar);
        $http.get('resume').success(function(data){
        });
        $scope.counter=0;
        updateProgressBar();
        $scope.form.counterRunning=false;
    };


    function startTimer(time){
        $scope.counter = time;
        updateProgressBar();
        $scope.form.counterRunning = true;
        clearInterval(intervalVar);
        intervalVar = setInterval(function(){
            $scope.$apply(function(){
                $scope.counter--;
                updateProgressBar();
                if( $scope.counter <= 0){
                    $scope.form.counterRunning = false;
                    clearInterval(intervalVar);
                }
            });
        }, 1000);
    }

    function updateProgressBar(){
        if($scope.counter <= 0)
            $scope.progressBar.width='100%';
        $scope.progressBar.width = Math.floor( $scope.counter /(startCounter / 100)) + '%';
    }
}
PauzerCtrl.$inject = ['$scope', '$http'];
