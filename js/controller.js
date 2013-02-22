'use strict';
function PauzerCtrl($scope, $http){

    var master = {
        limit: 20,
        time: 1,
        counterRunning: false,
    },
    intervalVar;

    $scope.form = angular.copy(master);

    $http.get('state').success(function(data){
        if(data.limit > 0){
            $scope.form = {
                limit: data.limit,
                time: data.time,
            };
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
            startTimer( $scope.form.time * 60);
        });
    };

    function clearTimer(){
        clearInterval(intervalVar);
        $http.get('resume').success(function(data){
        });
        $scope.counter=0;
        $scope.form.counterRunning=false;
    };


    function startTimer(time){
        $scope.counter = time;
        $scope.form.counterRunning = true;
        clearInterval(intervalVar);
        intervalVar = setInterval(function(){
            $scope.$apply(function(){
                $scope.counter--;
                if( $scope.counter <= 0){
                    $scope.form.counterRunning = false;
                    clearInterval(intervalVar);
                }
            });
        }, 1000);
    }


}
PauzerCtrl.$inject = ['$scope', '$http'];
