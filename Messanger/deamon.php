<?php

if(isset($_POST['attacked_server_name']) ){
    require_once 'sms.ru.php';

    $smsru = new SMSRU('api_id'); // Ваш уникальный программный ключ, который можно получить на главной странице

    $data = new stdClass();
    $data->to = '+7800553535';
    $data->text = 'Server: '.$_POST['attacked_server_name'].' were attacked;'; // Текст сообщения
    $sms = $smsru->send_one($data); // Отправка сообщения и возврат данных в переменную
}

