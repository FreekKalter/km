'use strict';
var kmApp = angular.module('kmApp', ['ngRoute', 'kmControllers' ,'kmAnimations' ,'ui.bootstrap' ]);

kmApp.config(['$routeProvider', function($routeProvider, $locationProvider){
    var d = new Date();
    $routeProvider.
    when('/input/:date', {
        templateUrl: 'partials/input.html',
        controller: 'kmInput'
    }).
    when('/overview/:category/:year/:month', {
        templateUrl:'partials/overview.html',
        controller: 'kmOverviewController'
    }).
    when('/overview/tijden', {
        redirectTo: '/overview/tijden/' + d.getFullYear() + '/' + (d.getMonth() +1)
    }).
    when('/overview', {
        redirectTo: '/overview/kilometers/' + d.getFullYear() + '/' + (d.getMonth() +1)
    }).
    otherwise({
        redirectTo: '/input/today'
    });
}]);
