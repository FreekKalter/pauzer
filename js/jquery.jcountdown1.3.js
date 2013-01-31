/* 
* jCountdown 1.3 jQuery Plugin
* Copyright 2011 Tom Ellis http://www.webmuse.co.uk
* Licensed under MIT License
* See http://www.webmuse.co.uk/license/
*/
(function ($) {

	$.fn.countdown = function( method /*,options*/ ) {  
	
		var defaults = {
				date: (new Date()),
				updateTime: 1000,
				htmlTemplate: "%{d} <span class=\"cd-time days\">days</span> %{h} <span class=\"cd-time hours\">hours</span> %{m} <span class=\"cd-time minutes\">mins</span> %{s} <span class=\"cd-time seconds\">sec</span>",
				minus: false,
				onChange: null,
				onComplete: null,
				onResume: null,
				onPause: null,
				leadingZero: false,
				offset: null,
				direction: 'down'
			},
			slice = [].slice,
			floor = Math.floor,
			msPerHour = 36E5,
			msPerDay = 864E5,
			rDate = /(%\{d\}|%\{h\}|%\{m\}|%\{s\})/g,
			rDays = /%\{d\}/,
			rHours = /%\{h\}/,
			rMins = /%\{m\}/,
			rSecs = /%\{s\}/,
			getTimezoneDate = function( offset ) {
			
				var hoursOffset = offset || 0,
					currentHours = 0,
					tempDate = new Date(),
					dateMS;
				
				hoursOffset = hoursOffset * msPerHour;
				currentHours = tempDate.getTime() - ( ( -tempDate.getTimezoneOffset() / 60 ) * msPerHour );
				dateMS = tempDate.setTime( currentHours + hoursOffset );
				
				return (new Date( dateMS ));
			},			
			timerFunc = function() {

				var $this = this,
					template,
					todaysDate,
					countdownDate,
					timeLeft,
					e_daysLeft,
					daysLeft,
					e_hrsLeft,
					hrsLeft,
					minsLeft,					
					e_minsleft,
					secLeft,
					time = "",
					settings = $this.data('jcdSettings');
					
				if( !settings ) {
					return;
				}
				
				template = settings.htmlTemplate;
				
				todaysDate = ( settings.offset === null ) ? new Date() : getTimezoneDate( settings.offset );
					
				countdownDate = new Date( settings.date );
				
				timeLeft = ( settings.direction === 'down' ) ? countdownDate.getTime() - todaysDate.getTime() :
					todaysDate.getTime() - countdownDate.getTime();
					
				e_daysLeft = timeLeft / msPerDay;
				daysLeft = floor( e_daysLeft );
				e_hrsLeft = ( e_daysLeft - daysLeft ) * 24;
				hrsLeft = floor( e_hrsLeft );
				minsLeft = floor( ( e_hrsLeft - hrsLeft ) * 60 );				
				e_minsleft = ( e_hrsLeft - hrsLeft ) * 60;
				secLeft = floor( (e_minsleft - minsLeft ) * 60 );
				
				if ( settings.leadingZero ) {
					daysLeft = zeroPad(daysLeft,2)
					hrsLeft = zeroPad(hrsLeft,2)
                    minsLeft = zeroPad(minsLeft, 2)
                    secLeft = zeroPad(secLeft,2)
				}

                if (daysLeft==0){
                    daysLeft = "";
                    if(hrsLeft==0){
                        hrsLeft=""
                    }
                }

				if ( settings.direction === 'down' && ( todaysDate <= countdownDate || settings.minus ) ) {
					time = template.replace( rDays, daysLeft ).replace( rHours, hrsLeft ).replace( rMins, minsLeft ).replace( rSecs, secLeft );
				} else if ( settings.direction === 'up' && ( countdownDate <= todaysDate || settings.minus ) ) {
					time = template.replace( rDays, daysLeft ).replace( rHours, hrsLeft ).replace( rMins, minsLeft ).replace( rSecs, secLeft );
				} else {
					time = template.replace( rDate, "00");
					settings.hasCompleted = true;
				}

                el = $('<p />').html(time);
                if( daysLeft == 0){
                    el.find('.days').remove();
                    if( hrsLeft == 0){
                        el.find('.hours').remove();
                    }
                }
                    
                settings.minsLeft = minsLeft
                settings.secLeft = secLeft
				$this.html( el );
				
				$this.trigger('change', [settings] );
				
				if ( settings.hasCompleted ){

					$this.trigger('complete.jcountdown');
					window.clearInterval( settings.timer );
				}
			},			
			methods = {
			
				init: function( options ){
					
					var opts = $.extend( {}, defaults, options ),
						template;
					
					template = opts.htmlTemplate;
					
					return this.each(function() {
						var $this = $(this),
							settings = {},
							todaysDate = ( opts.offset === null ) ? new Date() : getTimezoneDate( opts.offset ),
							countdownDate = new Date( opts.date ),
							timeLeft = ( opts.direction === 'down' ) ? countdownDate.getTime() - todaysDate.getTime() :
							todaysDate.getTime() - countdownDate.getTime(),
							e_daysLeft = timeLeft / msPerDay,
							daysLeft = floor(e_daysLeft),
							e_hrsLeft = (e_daysLeft - daysLeft) * 24, //Gets remainder and * 24
							hrsLeft = floor(e_hrsLeft),
							minsLeft = floor((e_hrsLeft - hrsLeft)*60),					
							e_minsleft = (e_hrsLeft - hrsLeft)*60, //Gets remainder and * 60
							secLeft = floor((e_minsleft - minsLeft)*60),
							time = "",
							func;

						if( opts.onChange ){
							$this.bind("change.jcountdown", opts.onChange );
						}
						
						if( opts.onComplete ){
							$this.bind("complete.jcountdown", opts.onComplete );
						}
						
						if( opts.onPause ){
							$this.bind("pause.jcountdown", opts.onPause );
						}

						if( opts.onResume ){
							$this.bind("resume.jcountdown", opts.onResume );
						}
						
                        if ( opts.leadingZero ) {
                            daysLeft = zeroPad(daysLeft,2)
                            hrsLeft = zeroPad(hrsLeft,2)
                            minsLeft = zeroPad(minsLeft, 2)
                            secLeft = zeroPad(secLeft,2)
                        }
			
						settings.hasCompleted = false;
						
                        if (daysLeft==0){
                            daysLeft = "";
                            if(hrsLeft==0){
                                hrsLeft=""
                            }
                        }
						//Set initial time
						if ( opts.direction === 'down' && ( todaysDate <= countdownDate || opts.minus ) ) {
							time = template.replace( rDays, daysLeft ).replace( rHours, hrsLeft ).replace( rMins, minsLeft ).replace( rSecs, secLeft );
						} else if ( opts.direction === 'up' && ( countdownDate <= todaysDate || opts.minus ) ) {
							time = template.replace( rDays, daysLeft ).replace( rHours, hrsLeft ).replace( rMins, minsLeft ).replace( rSecs, secLeft );
						} else {
							time = template.replace( rDate, "00");
							settings.hasCompleted = true;
						}

                        el = $('<p />').html(time);
                        if( daysLeft == 0){
                            el.find('.days').remove();
                            if( hrsLeft == 0){
                                el.find('.hours').remove();
                            }
                        }
						//Store settings so they can be accessed later
						
						settings.originalContent = $(this).html();
						settings.date = opts.date;
						settings.leadingZero = opts.leadingZero;
						settings.updateTime = opts.updateTime;
						settings.direction = opts.direction;
						settings.template = opts.htmlTemplate;
						settings.htmlTemplate = opts.htmlTemplate;
						settings.minus = opts.minus;
						settings.offset = opts.offset;
						settings.onChange = opts.onChange;
						settings.onComplete = opts.onComplete;
						settings.onResume = opts.onResume;
						settings.onPause = opts.onPause;
						
						if( !settings.hasCompleted ) {
							func = $.proxy( timerFunc, $this );
							settings.timer = window.setInterval( func, settings.updateTime );
						}
						
						$this.data( 'jcdSettings', settings );
						
                        settings.minsLeft = minsLeft
                        settings.secLeft = secLeft
						$this.html( el );
						
						if ( settings.hasCompleted ) {
						
							$this.trigger('complete.jcountdown');
							window.clearInterval( settings.timer );
						}
						
					});				
				
				},
				changeSettings: function( options ) {

					//Like resume but with resetting/changing options
					
					return this.each(function() {
						var $this  = $(this),
							settings,
							template,
							todaysDate,
							countdownDate,
							timeLeft,
							e_daysLeft,
							daysLeft,
							e_hrsLeft,
							hrsLeft,
							minsLeft,					
							e_minsleft,
							secLeft,
							time = "",
							func;
							
						if( !$this.data('jcdSettings') ) {
							return true;
						}
						
						settings = $.extend( {}, $this.data('jcdSettings'), options );
						
						template = settings.htmlTemplate;

						todaysDate = ( settings.offset === null ) ? new Date() : getTimezoneDate( settings.offset );
						countdownDate = new Date( settings.date );						
						timeLeft = ( settings.direction === 'down' ) ? countdownDate.getTime() - todaysDate.getTime() :
						todaysDate.getTime() - countdownDate.getTime();
						e_daysLeft = timeLeft / msPerDay;
						daysLeft = floor( e_daysLeft );
						e_hrsLeft = ( e_daysLeft - daysLeft ) * 24; //Gets remainder and * 24
						hrsLeft = floor( e_hrsLeft );
						minsLeft = floor( ( e_hrsLeft - hrsLeft ) * 60 );					
						e_minsleft = ( e_hrsLeft - hrsLeft ) * 60; //Gets remainder and * 60
						secLeft = floor( ( e_minsleft - minsLeft ) * 60);
						
						$this.unbind('.jcountdown');
						
						window.clearInterval( settings.timer );
						
						if( settings.onChange ) {
							$this.bind('change.jcountdown', settings.onChange);
						}

						if( settings.onComplete ) {
							$this.bind('complete.jcountdown', settings.onComplete);
						}
						
						if( settings.onPause ){
							$this.bind("pause.jcountdown", settings.onPause );
						}

						if( settings.onResume ){
							$this.bind("resume.jcountdown", settings.onResume );
						}
						
						if ( settings.direction === 'down' && ( todaysDate <= countdownDate || settings.minus ) ) {
							time = template.replace( rDays, daysLeft ).replace( rHours, hrsLeft ).replace( rMins, minsLeft ).replace( rSecs, secLeft );
						} else if ( settings.direction === 'up' && ( countdownDate <= todaysDate || settings.minus ) ) {
							time = template.replace( rDays, daysLeft ).replace( rHours, hrsLeft ).replace( rMins, minsLeft ).replace( rSecs, secLeft );
						} else {
							time = template.replace( rDate, "00");
							settings.hasCompleted = true;
						}

						func = $.proxy( timerFunc, $this );

						settings.timer =  window.setInterval( func, settings.updateTime );
						
						$this.data('jcdSettings', settings);
						
						if ( settings.hasCompleted ) {
							$this.trigger('complete.jcountdown');
							window.clearInterval( settings.timer );
						}
														
					});				
				
				},
				resume: function() {
				
					return this.each(function() {
						var $this = $(this),
							settings,
							template,
							func,
							todaysDate,
							countdownDate,
							timeLeft,
							e_daysLeft,
							daysLeft,
							e_hrsLeft,
							hrsLeft,
							minsLeft,					
							e_minsleft,
							secLeft,
							time = "";
							
						settings = $this.data('jcdSettings');
						
						if( !settings ){
							return true;
						}
						
						func = $.proxy( timerFunc, $this );
						
						template = settings.htmlTemplate;

						todaysDate = ( settings.offset === null ) ? new Date() : getTimezoneDate( settings.offset );
						countdownDate = new Date( settings.date );						
						timeLeft = ( settings.direction === 'down' ) ? countdownDate.getTime() - todaysDate.getTime() :
							todaysDate.getTime() - countdownDate.getTime();
						e_daysLeft = timeLeft / msPerDay;
						daysLeft = floor( e_daysLeft );
						e_hrsLeft = ( e_daysLeft - daysLeft ) * 24;
						hrsLeft = floor( e_hrsLeft );
						minsLeft = floor( ( e_hrsLeft - hrsLeft ) * 60 );					
						e_minsleft = ( e_hrsLeft - hrsLeft ) * 60;
						secLeft = floor( ( e_minsleft - minsLeft ) * 60 );

						if ( settings.direction === 'down' && ( todaysDate <= countdownDate || settings.minus ) ) {
							time = template.replace( rDays, daysLeft ).replace( rHours, hrsLeft ).replace( rMins, minsLeft ).replace( rSecs, secLeft );
						} else if ( settings.direction === 'up' && ( countdownDate <= todaysDate || settings.minus ) ) {
							time = template.replace( rDays, daysLeft ).replace( rHours, hrsLeft ).replace( rMins, minsLeft ).replace( rSecs, secLeft );
						} else {
							time = template.replace( rDate, "00" );
							settings.hasCompleted = true;
						}
						
						settings.timer =  window.setInterval( func, settings.updateTime );
						
						$this.data('jcdSettings', settings);
						
						$this.trigger('resume.jcountdown');
						
						$this.html( time ).trigger('change.jcountdown');
						
						if ( settings.hasCompleted ) {
						
							$this.trigger('complete.jcountdown');
							window.clearInterval( settings.timer );
						}	
					});
				},
				pause: function() {
				
					return this.each(function() {
						var $this = $(this),
							settings = $.data( $this[0], 'jcdSettings' );

						if( !settings ){
							return true;
						}
						
						$this.trigger('pause.jcountdown');	
						window.clearInterval( settings.timer );
					});
				},
				complete: function() {

					return this.each(function() {
						var $this = $(this),
							settings = $this.data('jcdSettings' );

						if( !settings ){
							return true;
						}						
						
						window.clearInterval( settings.timer );
						settings.hasCompleted = true;
						
						$this.data('jcdSettings', settings);						
						$this.trigger('complete.jcountdown');
					});		
				},
				destroy: function(){
				
					return this.each(function() {
						var $this = $(this),
							settings;
						
						settings = $this.data( 'jcdSettings' );
						
						if( !settings ){
							return true;
						}
						
						$this.unbind('.jcountdown');
						$this.html( settings.originalContent );
						$this.removeData('jcdSettings');
					});
				},
				getSettings: function( name ){
				
					var settings,
						$this = $(this[0]);
					
					settings = $this.data( 'jcdSettings' );
					
					if( !settings ){
						return undefined;
					}
					
					if( name ) {
						
						if( settings[name] ) {
							return settings[name];
						}
						return undefined;
					}
						
					return settings;
				}
			};
		
		if( methods[method] ) {

			return methods[method].apply( this, slice.call( arguments, 1 ) );
		
		} else if( $.type( method ) === 'object' || !method ) {
		
			return methods.init.apply( this, arguments );
		} else {
			
			$.error('Method '+ method+' does not exist in the jCountdown Plugin');
		}
	};
       
})(jQuery);

function zeroPad(num, places) {
  var zero = places - num.toString().length + 1;
  return Array(+(zero > 0 && zero)).join("0") + num;
}
