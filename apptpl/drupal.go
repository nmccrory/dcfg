package apptpl

import (
	"os"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/drud/dcfg/dcfglib"
)

// DrupalConfig encapsulates all the configurations for a Drupal site.
type DrupalConfig struct {
	ConfigSyncDir    string
	DatabaseName     string
	DatabaseUsername string
	DatabasePassword string
	DatabaseHost     string
	DatabaseDriver   string
	DatabasePort     int
	DatabasePrefix   string
	HashSalt         string
	Hostname         string
	IsDrupal8        bool
}

// NewDrupalConfig produces a DrupalConfig object with default.
func NewDrupalConfig() *DrupalConfig {
	return &DrupalConfig{
		ConfigSyncDir:    "/var/www/html/sync",
		DatabaseName:     "data",
		DatabaseUsername: "root",
		DatabasePassword: "root",
		DatabaseHost:     "127.0.0.1",
		DatabaseDriver:   "mysql",
		DatabasePort:     3306,
		DatabasePrefix:   "",
		HashSalt:         dcfglib.PassTheSalt(),
		IsDrupal8:        false,
	}
}

const (
	drupalTemplate = `<?php
{{ $config := . }}
/* Automatically generated Drupal settings.php file. */

$databases['default']['default'] = array(
  'database' => "{{ $config.DatabaseName }}",
  'username' => "{{ $config.DatabaseUsername }}",
  'password' => "{{ $config.DatabasePassword }}",
  'host' => "{{ $config.DatabaseHost }}",
  'driver' => "{{ $config.DatabaseDriver }}",
  'port' => {{ $config.DatabasePort }},
  'prefix' => "{{ $config.DatabasePrefix }}",
);

ini_set('session.gc_probability', 1);
ini_set('session.gc_divisor', 100);
ini_set('session.gc_maxlifetime', 200000);
ini_set('session.cookie_lifetime', 2000000);

{{ if $config.IsDrupal8 }}

$settings['hash_salt'] = '{{ $config.HashSalt }}';

$settings['file_scan_ignore_directories'] = [
  'node_modules',
  'bower_components',
];

 $config_directories = array(
   CONFIG_SYNC_DIRECTORY => '{{ config.ConfigSyncDir }}',
 );


{{ else }}

$drupal_hash_salt = '{{ $config.HashSalt }}';
$base_url = '{{ $config.DeployURL }}';

if (isset($_SERVER['HTTP_X_FORWARDED_PROTO']) &&
  $_SERVER['HTTP_X_FORWARDED_PROTO'] == 'https') {
  $_SERVER['HTTPS'] = 'on';
}
{{ end }}

if (file_exists(__DIR__ . '/custom.settings.php')) {
  include __DIR__ . '/custom.settings.php';
}

if (isset($_ENV['DEPLOY_NAME']) && $_ENV['DEPLOY_NAME'] == 'local' && file_exists(__DIR__ . '/settings.local.php')) {
  include __DIR__ . '/settings.local.php';
}


// This is super ugly but it determines whether or not drush should include a custom settings file which allows
// it to work both within a docker container and natively on the host system.
if (!empty($_SERVER["argv"]) && strpos($_SERVER["argv"][0], "drush") && empty($_ENV['DEPLOY_NAME'])) {
  include __DIR__ . '../../../../drush.settings.php';
}
`
)

// WriteDrupalConfig dynamically produces valid settings.php file by combining a configuration
// object with a data-driven template.
func WriteDrupalConfig(drupalConfig *DrupalConfig, filePath string) error {
	tmpl, err := template.New("drupalConfig").Funcs(sprig.TxtFuncMap()).Parse(drupalTemplate)
	if err != nil {
		return err
	}
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	err = tmpl.Execute(file, drupalConfig)
	if err != nil {
		return err
	}
	return nil
}
